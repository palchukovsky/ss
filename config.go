// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

////////////////////////////////////////////////////////////////////////////////

const (
	// LambdaMaxRunTime is the max lambda run time before a forced kill.
	// It has leeway to complete all service things.
	LambdaMaxRunTime = 2501 * time.Millisecond
	// LambdaMaxRunTimeInclusive is the max inclusive lambda run time before
	// a forced kill.
	LambdaMaxRunTimeInclusive = 1501 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////
type Config struct {
	SS struct {
		Service ServiceConfig `json:"service"`
		Build   Build         `json:"build"`
		Log     logConfig     `json:"log"`
	} `json:"ss"`
}

type ConfigWrapper interface {
	GetBasePtr() interface{}
	GetSSPtr() *Config
}

////////////////////////////////////////////////////////////////////////////////

type ServiceConfig struct {
	Endpoint string `json:"endpoint"`
	AWS      struct {
		AccountID string       `json:"accountId"`
		Region    string       `json:"region"`
		AccessKey AWSAccessKey `json:"accessKey"`
		Gateway   struct {
			App struct {
				ID string `json:"id"`
				// Endpoint is the app API gateway full path, set by builder.
				Endpoint string `json:"endpoint"`
			} `json:"app"`
			Auth struct {
				ID string `json:"id"`
			} `json:"auth"`
		} `json:"gateway"`
	} `json:"aws"`
	Firebase   FirebaseConfig `json:"firebase"`
	PrivateKey struct {
		RSA RSAPrivateKey `json:"rsa"`
	} `json:"privateKey"`
}

func (ServiceConfig) IsExtraLogEnabled() bool { return !S.Build().IsProd() }

////////////////////////////////////////////////////////////////////////////////

type Build struct {
	Version    string `json:"version"`    // verbose product version
	Commit     string `json:"commit"`     // full repository commid ID
	ID         string `json:"id"`         // verbose shot build ID to compare
	Builder    string `json:"builder"`    // build ID on builder
	Maintainer string `json:"maintainer"` // person who started build
}

// IsProd returns true if build is production.
func (build Build) IsProd() bool { return build.Version != "dev" }

////////////////////////////////////////////////////////////////////////////////

type logConfig struct {
	Sentry     string `json:"sentry,omitempty"`
	Loggly     string `json:"loggly,omitempty"`
	Logzio     string `json:"logzio,omitempty"`
	Papertrail string `json:"papertrail,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

type awsAccessKey struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type AWSAccessKey struct {
	IsUsed bool
	awsAccessKey
}

func (key *AWSAccessKey) UnmarshalJSON(source []byte) error {
	if !key.IsUsed {
		key.awsAccessKey = awsAccessKey{}
		return nil
	}
	var value awsAccessKey
	if err := json.Unmarshal(source, &value); err != nil {
		return err
	}
	key.awsAccessKey = value
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type FirebaseConfig struct {
	IsUsed bool
	value  []byte
}

func (config FirebaseConfig) GetJSON() []byte {
	if config.value == nil {
		S.Log().Panic(NewLogMsg(`Firebase config is not set`))
	}
	return config.value
}

func (config *FirebaseConfig) UnmarshalJSON(source []byte) error {
	if !config.IsUsed {
		config.value = nil
		return nil
	}
	var value json.RawMessage
	if err := json.Unmarshal(source, &value); err != nil {
		return err
	}
	config.value = value
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type RSAPrivateKey struct {
	IsUsed bool
	value  *rsa.PrivateKey
}

func (key RSAPrivateKey) Get() *rsa.PrivateKey {
	if key.value == nil {
		S.Log().Panic(NewLogMsg(`RSA key is not set`))
	}
	return key.value
}

func (key *RSAPrivateKey) UnmarshalJSON(jsonSource []byte) error {
	if !key.IsUsed {
		key.value = nil
		return nil
	}
	var source string
	if err := json.Unmarshal(jsonSource, &source); err != nil {
		return fmt.Errorf(`failed to unmarshal key string: "%w"`, err)
	}
	bin, err := base64.RawStdEncoding.DecodeString(source)
	if err != nil {
		return fmt.Errorf(`failed to decode Base64 with RSA private key: "%w"`, err)
	}
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(bin)
	if err != nil {
		return fmt.Errorf(`failed to parse PKCS1 Private Key: "%w"`, err)
	}
	key.value = rsaPrivateKey
	return nil
}

////////////////////////////////////////////////////////////////////////////////
