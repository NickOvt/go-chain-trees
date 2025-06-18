package utils

import (
	"crypto/sha3"

	"github.com/fxamacker/cbor/v2"
)

type CBORData []byte
type Hash []byte

// Encode any given value to CBOR
func EncodeCBOR(v any) (CBORData, error) {
	b, err := cbor.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// Decode CBOR value to data
func DecodeCBOR[T any](cborData CBORData) (T, error) {
	var output T
	err := cbor.Unmarshal(cborData, &output)

	if err != nil {
		return output, err
	}

	return output, nil
}

// Generate hash of given bytes
func GenerateHash(data []byte) Hash {
	hash := sha3.Sum384(data)
	return hash[:]
}

func ConcatDataAndGenerateHash(data ...[]byte) Hash {
	// Calculate total length needed to preallocate memory
	totalLen := 0
	for _, slice := range data {
		totalLen += len(slice)
	}

	combined := make([]byte, 0, totalLen)

	for _, slice := range data {
		combined = append(combined, slice...)
	}

	// Use your existing generateHash function
	return GenerateHash(combined)
}
