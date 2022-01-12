// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

// Answers "200 OK" with backend information at success, or "406 Not Acceptable"
// if the app client version has to be upgraded.

package authlambda

import (
	"context"
	"time"

	"firebase.google.com/go/auth"

	ss "github.com/palchukovsky/ss"
	apigateway "github.com/palchukovsky/ss/api/gateway"
	apiauth "github.com/palchukovsky/ss/api/gateway/auth"
	ssdb "github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	rest "github.com/palchukovsky/ss/lambda/gateway/rest"
)

type request = string // token ID

type response struct {
	User        responseUser        `json:"user"`
	AccessToken apiauth.AccessToken `json:"token"`
	Backend     responseBackend     `json:"back"`
}

type responseUser struct {
	ID          ss.UserID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone,omitempty"`
	PhotoURL    string    `json:"photoUrl,omitempty"`
	IsNew       *bool     `json:"new,omitempty"`
}

type responseBackend struct {
	Endpoint string `json:"endpoint"`
	Build    string `json:"build"`
	Version  string `json:"ver"`
}

func newResponse(
	user FirebaseIndex,
	accessToken apiauth.AccessToken,
	isNew bool,
) response {
	build := ss.S.Build()
	result := response{
		User: responseUser{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			PhotoURL:    user.PhotoURL,
		},
		AccessToken: accessToken,
		Backend: responseBackend{
			Endpoint: ss.S.Config().Endpoint,
			Build:    build.ID,
			Version:  build.Version,
		},
	}
	if isNew {
		result.User.IsNew = &isNew
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

type Policy interface {
	CheckNewUserName(newUserDisplayName string) string
	CheckCreateUserTans(trans ddb.WriteTrans, user ss.UserID, isAnonymous bool)
}

////////////////////////////////////////////////////////////////////////////////

func Init(
	initService func(projectPackage string, params ss.ServiceParams),
	policy Policy,
) {
	apiauth.Init(
		func() rest.Lambda {
			result := lambda{
				db:     ddb.GetClientInstance(),
				policy: policy,
			}
			var err error
			result.firebase, err = ss.S.Firebase().Auth(context.Background())
			if err != nil {
				ss.S.Log().Panic(ss.NewLogMsg(`failed auth Firebase`).AddErr(err))
			}
			return result
		},
		func(projectPackage string) {
			initService(
				projectPackage,
				ss.ServiceParams{
					IsAuth:     true,
					IsFirebase: true,
				})
		})
}

func Run() { apiauth.Run() }

////////////////////////////////////////////////////////////////////////////////

type lambda struct {
	db       ddb.Client
	firebase *auth.Client
	policy   Policy
}

func (lambda lambda) Execute(request rest.Request) error {

	if ok, err := lambda.CheckClientVersionActuality(request); err != nil {
		return err
	} else if !ok {
		request.RespondWithNotAcceptable()
		return nil
	}

	token, firebaseUser, isAnonymous := lambda.getFirebaseUser(request)
	if token == nil || firebaseUser == nil {
		return nil
	}

	user := lambda.findUser(firebaseUser.UID)
	isNew := user == nil
	if user == nil {
		user = &FirebaseIndex{}
	}

	if isNew {
		isNew = lambda.createUserUser(*firebaseUser, user, isAnonymous, request)
		if isNew {
			request.Log().Info(
				ss.
					NewLogMsg(
						"new user added from Firebase, access expires after %d mins",
						int(time.Unix(token.Expires, 0).Sub(time.Now().UTC()).Minutes())).
					Add(user.ID).
					AddVal("firebaseSignInProvider", token.Firebase.SignInProvider))
		}
	}

	if !isNew {
		if !user.IsAnonymous() && isAnonymous {
			request.Log().Panic(
				ss.
					NewLogMsg("record was is not anonymous, but now it is").
					AddDump(*user))
		}
		request.Log().Debug(
			ss.
				NewLogMsg(
					"user authed by Firebase, access expires after %d mins",
					int(time.Unix(token.Expires, 0).Sub(time.Now().UTC()).Minutes())).
				Add(user.ID).
				AddVal("firebaseSignInProvider", token.Firebase.SignInProvider))
		lambda.updateUser(
			*firebaseUser,
			request,
			user,
			ss.BoolPtrIfSet(isAnonymous))
	}

	accessToken, err := apiauth.NewAccessToken(user.ID, token.Expires)
	if err != nil {
		request.Log().Panic(
			ss.NewLogMsg(`failed to create access token`).AddErr(err))
	}

	request.Respond(newResponse(*user, accessToken, isNew))
	return nil
}

func (lambda lambda) CheckClientVersionActuality(
	lambdaRequest rest.Request,
) (
	bool,
	error,
) {
	return apigateway.CheckClientVersionActuality(
		lambdaRequest.ReadClientVersion())
}

func (lambda lambda) getFirebaseUser(
	lambdaRequest rest.Request,
) (
	token *auth.Token,
	user *auth.UserRecord,
	isAnonymous bool,
) {

	var tokenID request
	lambdaRequest.ReadRequest(&tokenID)

	var err error
	token, err = lambda.firebase.VerifyIDTokenAndCheckRevoked(
		lambdaRequest.GetContext(),
		tokenID)
	if err != nil {
		lambdaRequest.Log().Panic(
			ss.NewLogMsg(`failed to verify token`).AddErr(err))
	}

	if token.UID == "" {
		lambdaRequest.Log().Panic(
			ss.NewLogMsg(`token is not authed by Firebase`).AddDump(*token))
	}

	user, err = lambda.firebase.GetUser(lambdaRequest.GetContext(), token.UID)
	if err != nil {
		lambdaRequest.Log().Panic(
			ss.NewLogMsg(`failed to get Firebase user`).AddDump(*token).AddErr(err))
	}
	if user == nil {
		lambdaRequest.Log().Panic(ss.NewLogMsg(`no Firebase user`).AddDump(*token))
		return // to prevent lint warning "go-golangci-lint"
	}

	if user.Disabled {
		lambdaRequest.Log().Panic(
			ss.NewLogMsg(`Firebase user is disabled`).AddDump(*token).AddDump(*user))
	}

	isAnonymous = token.Firebase.SignInProvider == "anonymous"

	if !isAnonymous && !user.EmailVerified {
		// Firebase marks an email as "verified" only if really knows that it is
		// verified, so it can't say it about 3rd parties such as Facebook.
		// To resolve, we need to send a confirmation email
		// for each 3rd party service.
		for len(user.ProviderUserInfo) != 1 {
			if user.ProviderUserInfo[0].ProviderID != `facebook.com` {
				lambdaRequest.Log().Panic(
					ss.
						NewLogMsg(`Firebase user email is not verified`).
						AddDump(*token).
						AddDump(*user))
			}
		}
	}

	return
}

func (lambda lambda) findUser(firebaseUserID string) *FirebaseIndex {
	var result FirebaseIndex
	isFound := lambda.
		db.
		Index(&result).
		Query("fId = :i", ddb.Values{":i": firebaseUserID}).
		RequestOne()
	if !isFound {
		return nil
	}
	return &result
}

func (lambda lambda) createUserUser(
	source auth.UserRecord,
	user *FirebaseIndex,
	isAnonymous bool,
	request rest.Request,
) bool {

	record, uniqueIndex := ssdb.NewFirebaseUser(
		source.UID,
		lambda.policy.CheckNewUserName(source.DisplayName),
		isAnonymous)
	record.Email = source.Email
	record.PhoneNumber = source.PhoneNumber
	record.PhotoURL = source.PhotoURL

	trans := ddb.NewWriteTrans(false)
	trans.CreateIfNotExists(record)
	trans.CreateIfNotExists(uniqueIndex).AllowConditionalCheckFail()
	lambda.policy.CheckCreateUserTans(trans, record.ID, isAnonymous)

	if !lambda.db.Write(trans).IsSuccess() {
		// Firebase ID already registered.
		return false
	}

	*user = NewFirebaseIndex(record)
	return true
}

type updateRecord struct {
	ssdb.UserRecord
	OriginalName string `json:"origName"`
	OwnName      string `json:"ownName,omitempty"`
	Email        string `json:"email,omitempty"`
	PhoneNumber  string `json:"phone,omitempty"`
	PhotoURL     string `json:"photoUrl,omitempty"`
}

func (r updateRecord) GetName() string {
	if r.OwnName != "" {
		return r.OwnName
	}
	return r.OriginalName
}

func (r *updateRecord) Clear() { *r = updateRecord{} }

func (lambda lambda) updateUser(
	userRecord auth.UserRecord,
	request rest.Request,
	user *FirebaseIndex,
	isAnonymous *bool,
) {
	update := lambda.db.Update(ssdb.NewUserKey(user.ID))

	if userRecord.DisplayName != "" {
		update.Set("origName = :n").Value(":n", userRecord.DisplayName)
	}

	if userRecord.PhotoURL != "" {
		update.Set("photoUrl = :u").Value(":u", userRecord.PhotoURL)
	} else {
		update.Remove("photoUrl")
	}

	if userRecord.Email != "" {
		update.Set("email = :e").Value(":e", userRecord.Email)
	} else {
		update.Remove("email")
	}

	if userRecord.PhoneNumber != "" {
		update.Set("phone = :p").Value(":p", userRecord.PhoneNumber)
	} else {
		update.Remove("phone")
	}

	if isAnonymous != nil {
		if ss.IsBoolSet(isAnonymous) {
			update.
				Set("anonymExpiration = :anonymExpiration").
				Value(
					":anonymExpiration",
					ssdb.NewUserAnonymousRecordExpirationTime(ss.Now()))
		} else {
			update.Remove("anonymExpiration")
		}
	}

	var record updateRecord
	if !update.RequestAndReturn(&record).IsSuccess() {
		request.Log().Error(
			ss.NewLogMsg(`failed to find update user to update`).Add(user.ID))
		return
	}

	user.Name = record.GetName()
	user.Email = record.Email
	user.PhoneNumber = record.PhoneNumber
	user.PhotoURL = record.PhotoURL
}
