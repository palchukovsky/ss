// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/crypto"
	"github.com/palchukovsky/ss/crypto/padding"
)

// Encrypt implements AES CBC with HMAC authenticated encryption
// encoded by Base64.
func Encrypt(
	source []byte,
	key crypto.AES128Key,
) (
	data string,
	auth string,
) {
	source = padding.AddPKCS7(source, 16)

	encrypted := make([]byte, aes.BlockSize+len(source)) // 1st block for IV

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		ss.S.Log().Panic(
			ss.NewLogMsg("failed to generate IV for AES packet").AddErr(err))
	}

	aesCipher, err := aes.NewCipher(key.Get())
	if err != nil {
		ss.S.Log().Panic(ss.NewLogMsg("failed to init AES cipher").AddErr(err))
	}

	cipher.
		NewCBCEncrypter(aesCipher, iv).
		CryptBlocks(encrypted[aes.BlockSize:], source)

	// Uses the same kye for HMAC as for encryption:
	//   With HMAC vs AES, no such interference is known. The general feeling of
	//   cryptographers is that AES and SHA-1 (or SHA-256) are "sufficiently
	//   different" that there should be no practical issue with using the same
	//   key for AES and HMAC/SHA-1.
	// https://crypto.stackexchange.com/questions/8081/using-the-same-secret-key-for-encryption-and-authentication-in-a-encrypt-then-ma/8086#8086
	hasher := hmac.New(sha256.New, key.Get())
	if _, err := hasher.Write(encrypted); err != nil {
		ss.S.Log().Panic(
			ss.NewLogMsg("failed to generate HMAC for AES").AddErr(err))
	}
	hash := hasher.Sum(nil) // len(hash) = 32

	data = base64.RawStdEncoding.EncodeToString(encrypted)
	auth = base64.RawStdEncoding.EncodeToString(hash)
	return
}
