// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package crypto

import (
	"encoding/base64"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////

const aes128KeyLen = 16

type AES128Key [aes128KeyLen]byte

func NewAES128KeyFromBase64(source string) (AES128Key, error) {
	var result AES128Key
	data, err := base64.RawStdEncoding.DecodeString(source)
	if err != nil {
		return result,
			fmt.Errorf(`failed to read AES128 key from Base64: "%w"`, err)
	}
	if len(data) != aes128KeyLen {
		return result, fmt.Errorf(`AES128 key has wrong length %d`, len(data))
	}
	copy(result[:], data)
	return result, nil
}

func (k AES128Key) Get() []byte { return k[:] }

////////////////////////////////////////////////////////////////////////////////
