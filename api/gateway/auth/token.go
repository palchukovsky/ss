// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apiauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/palchukovsky/ss"
)

func newToken(source interface{}) (string, error) {
	creator := newTokenCreator()
	data, err := creator.SerializeValue(source)
	if err != nil {
		return "", err
	}
	signature, err := creator.Sign(data)
	if err != nil {
		return "", err
	}
	data = append(data, 0)
	data = append(data, signature...)
	return base64.RawStdEncoding.EncodeToString(data), nil
}

func parseToken(source string, result interface{}) (error, error) {
	data, err := base64.RawStdEncoding.DecodeString(source)
	if err != nil {
		return err, nil
	}
	reader := newTokenReader()
	data, signature, err := reader.GetValueAndSignature(data)
	if err != nil {
		return err, nil
	}
	if err := reader.VerifySignature(data, signature); err != nil {
		return nil, err
	}
	if err := reader.ParseValue(data, result); err != nil {
		return nil, err
	}
	return nil, nil
}

func parseTimeToken(source string) (time.Time, error) {
	result, err := strconv.ParseInt(source, 16, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf(`field to parse time token %q: "%w"`,
			source, err)
	}
	return time.Unix(result, 0), nil
}

////////////////////////////////////////////////////////////////////////////////

type tokenCreator struct{}

func newTokenCreator() tokenCreator { return tokenCreator{} }

func (tokenCreator) SerializeValue(value interface{}) ([]byte, error) {
	result, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf(`failed to serialize: "%w"`, err)
	}
	return result, nil
}

func (tokenCreator) Sign(data []byte) ([]byte, error) {
	hash := crypto.SHA256
	hasher := hash.New()
	if _, err := hasher.Write(data); err != nil {
		return nil, fmt.Errorf(`failed to calc hash for signature: "%w"`, err)
	}
	result, err := rsa.SignPSS(
		rand.Reader,
		ss.S.Config().PrivateKey.RSA.Get(),
		hash,
		hasher.Sum(nil),
		&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
	if err != nil {
		return nil, fmt.Errorf(`failed to sign: "%w"`, err)
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

type tokenReader struct{}

func newTokenReader() tokenReader { return tokenReader{} }

func (tokenReader) ParseValue(data []byte, result interface{}) error {
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf(`failed to parse: "%w"`, err)
	}
	return nil
}

func (tokenReader) GetValueAndSignature(
	data []byte) (value []byte, signature []byte, err error) {
	for i, v := range data {
		if v != 0 {
			continue
		}
		return data[:i], data[i+1:], nil
	}
	return nil, nil, errors.New("wrong format")
}

func (tokenReader) VerifySignature(data []byte, signature []byte) error {
	hash := crypto.SHA256
	hasher := hash.New()
	if _, err := hasher.Write(data); err != nil {
		return fmt.Errorf(`failed to calc hash to verify signature: "%w"`, err)
	}
	err := rsa.VerifyPSS(
		&ss.S.Config().PrivateKey.RSA.Get().PublicKey,
		hash,
		hasher.Sum(nil),
		signature, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto})
	if err != nil {
		return fmt.Errorf(`failed to verify signature: "%w"`, err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
