// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package authlambda

import (
	"context"
	"fmt"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"

	petname "github.com/dustinkirkland/golang-petname"
	ss "github.com/palchukovsky/ss"
	apiauth "github.com/palchukovsky/ss/api/gateway/auth"
	ssdb "github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	rest "github.com/palchukovsky/ss/lambda/gateway/rest"
	"google.golang.org/api/option"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	apiauth.Init(
		func() rest.Lambda {
			result := lambda{db: ddb.GetClientInstance()}
			options := option.WithCredentialsJSON(
				ss.S.Config().Firebase.GetJSON())
			firebase, err := firebase.NewApp(context.Background(), nil, options)
			if err != nil {
				ss.S.Log().Panic(ss.NewLogMsg(`failed to init Firebase`).AddErr(err))
			}
			result.firebase, err = firebase.Auth(context.Background())
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
}

type responseBackend struct {
	Endpoint string `json:"endpoint"`
	Build    string `json:"build"`
	Version  string `json:"ver"`
}

func newResponse(
	user FirebaseIndex,
	accessToken apiauth.AccessToken,
) response {
	build := ss.S.Build()
	return response{
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
}

////////////////////////////////////////////////////////////////////////////////

type lambda struct {
	db       ddb.Client
	firebase *auth.Client
}

func (lambda lambda) Execute(request rest.Request) error {

	token, firebaseUser, err := lambda.getFirebaseUser(request)
	if err != nil || token == nil || firebaseUser == nil {
		return err
	}

	user, err := lambda.findUser(firebaseUser.UID)
	if err != nil {
		return fmt.Errorf("filed to find user by Firebase ID %q", token.UID)
	}

	isNew := user == nil
	if user == nil {
		user = &FirebaseIndex{}
	}

	if isNew {
		isNew, err = lambda.createUserUser(*firebaseUser, user)
		if err != nil {
			return err
		}
		if isNew {
			request.Log().Info(
				ss.NewLogMsg(
					"new user added from Firebase, access expires after %d mins",
					int(time.Unix(token.Expires, 0).Sub(time.Now().UTC()).Minutes())).
					AddUser(user.ID).
					AddVal("firebaseSignInProvider", token.Firebase.SignInProvider))
		}
	}

	if !isNew {
		request.Log().Debug(
			ss.NewLogMsg(
				"user authed by Firebase, access expires after %d mins",
				int(time.Unix(token.Expires, 0).Sub(time.Now().UTC()).Minutes())).
				AddUser(user.ID).
				AddVal("firebaseSignInProvider", token.Firebase.SignInProvider))
		lambda.updateUser(*firebaseUser, request, user)
	}

	accessToken, err := apiauth.NewAccessToken(user.ID, token.Expires)
	if err != nil {
		return fmt.Errorf(`failed to create access token: "%w"`, err)
	}

	request.Respond(newResponse(*user, accessToken))
	return nil
}

func (lambda lambda) getFirebaseUser(lambdaRequest rest.Request,
) (*auth.Token, *auth.UserRecord, error) {

	var tokenID request
	if err := lambdaRequest.ReadRequest(&tokenID); err != nil {
		return nil, nil, err
	}

	token, err := lambda.firebase.VerifyIDTokenAndCheckRevoked(
		lambdaRequest.GetContext(), tokenID)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to verify: "%w"`, err)
	}

	if token.UID == "" {
		return token, nil, fmt.Errorf("token is not authed by Firebase")
	}

	user, err := lambda.firebase.GetUser(lambdaRequest.GetContext(), token.UID)
	if err != nil {
		return token, nil, fmt.Errorf(`failed to get Firebase user %q: "%w"`,
			token.UID, err)
	}
	if user == nil {
		return token, nil, fmt.Errorf("no Firebase user %q", token.UID)
	}

	if user.Disabled {
		return token, user, fmt.Errorf("Firebase user %q is disabled.",
			token.UID)
	}
	if !user.EmailVerified {
		return token, user, fmt.Errorf("Firebase user %q email is not viridied.",
			token.UID)
	}

	return token, user, nil
}

func (lambda lambda) findUser(
	firebaseUserID string,
) (*FirebaseIndex, error) {
	var result FirebaseIndex
	isFound, err := lambda.
		db.
		Index(&result).
		Query("fId = :i", ddb.Values{":i": firebaseUserID}).
		RequestOne()
	if err != nil || !isFound {
		return nil, err
	}
	return &result, nil
}

func (lambda lambda) createUserUser(
	source auth.UserRecord,
	user *FirebaseIndex,
) (bool, error) {

	name := source.DisplayName
	if name == "" {
		name = petname.Generate(2, " ")
	}

	record, uniqueIndex := ssdb.NewFirebaseUser(source.UID, name)
	record.Email = source.Email
	record.PhoneNumber = source.PhoneNumber
	record.PhotoURL = source.PhotoURL

	trans := ddb.NewWriteTrans()
	trans.Create(record)      // 0
	trans.Create(uniqueIndex) // 1

	err := lambda.db.Write(trans)
	if err == nil {
		*user = NewFirebaseIndex(record)
		return true, nil
	}

	if ddb.ParseErrorConditionalCheckFailed(err, 1, 1) != nil {
		// Firebase ID already registered.
		return false, nil
	}

	// Unexpected error.
	return false, err
}

type updateRecord struct {
	ssdb.UserRecord
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone,omitempty"`
	PhotoURL    string `json:"photoUrl,omitempty"`
}

func (r *updateRecord) Clear() { *r = updateRecord{} }

func (lambda lambda) updateUser(
	userRecord auth.UserRecord,
	request rest.Request,
	user *FirebaseIndex,
) {

	sets := []string{}
	removes := []string{}
	values := ddb.Values{}

	if userRecord.DisplayName != "" {
		sets = append(sets, "name = :n")
		values[":n"] = userRecord.DisplayName
	}

	if userRecord.PhotoURL != "" {
		sets = append(sets, "photoUrl = :u")
		values[":u"] = userRecord.PhotoURL
	} else {
		removes = append(removes, "photoUrl")
	}

	if userRecord.Email != "" {
		sets = append(sets, "email = :e")
		values[":e"] = userRecord.Email
	} else {
		removes = append(removes, "email")
	}

	if userRecord.PhoneNumber != "" {
		sets = append(sets, "phone = :p")
		values[":p"] = userRecord.PhoneNumber
	} else {
		removes = append(removes, "phone")
	}

	var update string
	if len(sets) > 0 {
		update += "set " + strings.Join(sets, ",")
	}
	if len(removes) > 0 {
		update += " remove " + strings.Join(removes, ",")
	}

	var record updateRecord
	isFound, err := lambda.
		db.
		Update(ssdb.NewUserKey(user.ID), update).
		Values(values).
		RequestAndReturn(&record)
	if err != nil || !isFound {
		request.Log().Error(
			ss.NewLogMsg(`failed to update user`).AddUser(user.ID).AddErr(err))
		return
	}

	user.Name = record.Name
	user.Email = record.Email
	user.PhoneNumber = record.PhoneNumber
	user.PhotoURL = record.PhotoURL
}
