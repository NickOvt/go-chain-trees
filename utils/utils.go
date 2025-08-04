package utils

import (
	"crypto/sha3"

	"github.com/fxamacker/cbor/v2"
)

type CBORData []byte
type Hash []byte

// EncodeCBOR Encode any given value to CBOR
func EncodeCBOR(v any) (CBORData, error) {
	b, err := cbor.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// DecodeCBOR Decode CBOR value to data
func DecodeCBOR[T any](cborData CBORData) (T, error) {
	var output T
	err := cbor.Unmarshal(cborData, &output)

	if err != nil {
		return output, err
	}

	return output, nil
}

// GenerateHash Generate hash of given bytes
func GenerateHash(data []byte) Hash {
	hash := sha3.Sum384(data)
	return hash[:]
}

// ConcatDataAndGenerateHash Given n CBOR byte arrays concat them and calculate hash
func ConcatDataAndGenerateHash(data ...CBORData) Hash {
	// Calculate total length needed to preallocate memory
	totalLen := 0
	for _, slice := range data {
		totalLen += len(slice)
	}

	combined := make([]byte, 0, totalLen)

	for _, slice := range data {
		combined = append(combined, slice...)
	}

	return GenerateHash(combined)
}

// EncodeCBORList Encode any given value to CBOR. Accepts multiple values
// If there is an error with any of the values then whole function
// returns an error
func EncodeCBORList(v ...any) ([]CBORData, error) {
	cborDataList := make([]CBORData, len(v))

	for index, value := range v {
		b, err := cbor.Marshal(value)

		if err != nil {
			return nil, err
		}

		cborDataList[index] = b
	}

	return cborDataList, nil
}
