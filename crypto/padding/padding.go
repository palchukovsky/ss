// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package padding

import (
	"bytes"
)

func AddPKCS7(source []byte, blockSize int) []byte {
	count := blockSize - (len(source) % blockSize)
	if count == 0 {
		count = blockSize
	}
	return append(
		source,
		bytes.Repeat([]byte{byte(count)}, count)...)
}

/*
func removePKCS7(source []byte, blockSize int) []byte {
	len := len(source)
	paddingCount := int(source[len-1])

	if paddingCount > blockSize || paddingCount <= 0 {
		// Data is not padded (or not padded correctly), return as is.
		return source
	}

	padding := source[len-paddingCount : len-1]

	for _, b := range padding {
		if int(b) != paddingCount {
			// Data is not padded (or not padded correctly), return as is.
			return source
		}
	}

	return source[:len-paddingCount]
}
*/
