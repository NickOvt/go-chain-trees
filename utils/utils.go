package utils

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"

	"github.com/fxamacker/cbor/v2"
)

// CBORData represents encoded CBOR data as a byte slice
type CBORData []byte

// Hash represents a cryptographic hash as a byte slice
type Hash []byte

// EncodeCBOR encodes any given value to CBOR format.
// It accepts any type that can be marshaled to CBOR.
//
// Parameters:
//   - v: The value to encode (can be any type)
//
// Returns:
//   - CBORData: The CBOR-encoded data as bytes
//   - error: An error if encoding fails, nil otherwise
func EncodeCBOR(v any) (CBORData, error) {
	b, err := cbor.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// DecodeCBOR decodes CBOR data into the specified type T.
// Uses Go generics.
//
// Type Parameters:
//   - T: The target type to decode into
//
// Parameters:
//   - cborData: The CBOR-encoded data to decode
//
// Returns:
//   - T: The decoded value of type T
//   - error: An error if decoding fails, nil otherwise
func DecodeCBOR[T any](cborData CBORData) (T, error) {
	var output T
	err := cbor.Unmarshal(cborData, &output)

	if err != nil {
		return output, err
	}

	return output, nil
}

// GenerateHashSha384 generates an SHA-384 hash of the given byte data.
//
// Parameters:
//   - data: The byte slice to hash
//
// Returns:
//   - Hash: The SHA-384 hash as a byte slice
func GenerateHashSha384(data []byte) Hash {
	hash := sha3.Sum384(data)
	return hash[:]
}

// GenerateHashSha256 generates a SHA2-256 hash of the given byte data.
//
// Parameters:
//   - data: The byte slice to hash
//
// Returns:
//   - Hash: The SHA2-256 hash as a byte slice
func GenerateHashSha256(data []byte) Hash {
	hash := sha256.Sum256(data)
	return hash[:]
}

// GenerateHashSha512 generates an SHA-512 hash of the given byte data.
//
// Parameters:
//   - data: The byte slice to hash
//
// Returns:
//   - Hash: The SHA-512 hash as a byte slice
func GenerateHashSha512(data []byte) Hash {
	hash := sha512.Sum512(data)
	return hash[:]
}

type HashAlgo string
type GenerateHashFunc func(data []byte) Hash

const (
	SHA256 HashAlgo = "sha256"
	SHA384 HashAlgo = "sha384"
	SHA512 HashAlgo = "sha512"
)

var hashAlgoToHashFuncMap = map[HashAlgo]GenerateHashFunc{
	SHA256: GenerateHashSha256,
	SHA384: GenerateHashSha384,
	SHA512: GenerateHashSha512,
}

var hashAlgoToHashAlgoOutputBitCount = map[HashAlgo]int{
	SHA256: 256,
	SHA384: 384,
	SHA512: 512,
}

// GenerateHash generates a hash of the given byte data using specified HashAlgo.
//
// Parameters:
//   - hashAlgo: The HashAlgo used (SHA256, SHA384, SHA512)
//   - data: The byte slice to hash
//
// Returns:
//   - Hash: The HashAlgo hash of the data as byte slice
func GenerateHash(hashAlgo HashAlgo, data []byte) Hash {
	if hashFunc, ok := hashAlgoToHashFuncMap[hashAlgo]; ok {
		return hashFunc(data)
	}

	// default to SHA256 if not found
	return GenerateHashSha256(data)
}

// GenerateNullHash generates a null hash (hash of nil) for specified hashAlgo.
//
// Parameters:
//   - hashAlgo: The HashAlgo used (SHA256, SHA384, SHA512)
//
// Returns:
//   - Hash: The HashAlgo hash of nil
func GenerateNullHash(hashAlgo HashAlgo) Hash {
	if hashFunc, ok := hashAlgoToHashFuncMap[hashAlgo]; ok {
		return hashFunc(nil)
	}

	// default to SHA256 if not found
	return GenerateHashSha256(nil)
}

// GetHashAlgoOutputBitCount outputs the output bit count of a specified hashAlgo. Defaults to 256 bits for SHA256
//
// Parameters:
//   - hashAlgo: The HashAlgo used (SHA256, SHA384, SHA512)
//
// Returns:
//   - int: The count of bits in the output hash
func GetHashAlgoOutputBitCount(hashAlgo HashAlgo) int {
	if hashAlgoOutputBitCount, ok := hashAlgoToHashAlgoOutputBitCount[hashAlgo]; ok {
		return hashAlgoOutputBitCount
	}

	return 256 // default SHA256
}

// ConcatDataAndGenerateHash concatenates multiple CBOR byte arrays
// and calculates their combined hash.
//
// Parameters:
//   - hashAlgo: The HashAlgo used (SHA256, SHA384, SHA512)
//   - data: Variable number of CBORData slices to concatenate and hash
//
// Returns:
//   - Hash: The hash of the concatenated data
func ConcatDataAndGenerateHash(hashAlgo HashAlgo, data ...CBORData) Hash {
	// Calculate total length needed to preallocate memory
	totalLen := 0
	for _, slice := range data {
		totalLen += len(slice)
	}

	combined := make([]byte, 0, totalLen)

	for _, slice := range data {
		combined = append(combined, slice...)
	}

	return GenerateHash(hashAlgo, combined)
}

// ConcatHashesAndGenerateHash concatenates multiple hashes
// and calculates their combined hash.
//
// Parameters:
//   - hashAlgo: The HashAlgo used (SHA256, SHA384, SHA512)
//   - hashes: Variable number of hashes to concatenate and hash
//
// Returns:
//   - Hash: The hash of the concatenated hashes
func ConcatHashesAndGenerateHash(hashAlgo HashAlgo, hashes ...Hash) Hash {
	// Calculate total length needed to preallocate memory
	totalLen := 0
	for _, slice := range hashes {
		totalLen += len(slice)
	}

	combined := make([]byte, 0, totalLen)

	for _, slice := range hashes {
		combined = append(combined, slice...)
	}

	return GenerateHash(hashAlgo, combined)
}

// EncodeCBORList encodes multiple values to CBOR format and returns them
// as a slice of CBORData. If encoding fails for any value, the entire
// operation fails and returns an error.
//
// Parameters:
//   - v: Variable number of values to encode (each can be any type)
//
// Returns:
//   - []CBORData: A slice containing the CBOR-encoded data for each input value
//   - error: An error if encoding fails for any value, nil otherwise
func EncodeCBORList(v ...any) ([]CBORData, error) {
	cborDataList := make([]CBORData, len(v))

	for index, value := range v {
		encodedData, err := EncodeCBOR(value)

		if err != nil {
			return nil, err
		}

		cborDataList[index] = encodedData
	}

	return cborDataList, nil
}

// FlipLastBit flips last bit of byteArray argument (array of bytes)
//
// Parameters:
//   - byteArray: Input byte array
//
// Returns:
//   - []byte: A new byteArray with last bit flipped
func FlipLastBit(byteArray []byte) []byte {
	byteArray[len(byteArray)-1] ^= 1
	return byteArray
}

// FlipBitAtN flips n-th bit of byteArray argument (array of bytes)
//
// Parameters:
//   - byteArray: Input byte array
//
// Returns:
//   - []byte: A new byteArray with n-th bit flipped
func FlipBitAtN(byteArray []byte, n int) []byte {
	byteIdx := n / 8
	bitIdx := 7 - (n % 8)
	byteArray[byteIdx] ^= 1 << bitIdx
	return byteArray
}
