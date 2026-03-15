package utils

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/binary"
	"hash"
	"io"

	"github.com/fxamacker/cbor/v2"
)

// CBORData represents encoded CBOR data as a byte slice
type CBORData []byte

// Hash represents a cryptographic hash as a byte slice
type Hash []byte

func mustDeterministicCBORMarshalMode() cbor.EncMode {
	mode, err := cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		panic(err)
	}

	return mode
}

var deterministicCBORMarshal = mustDeterministicCBORMarshalMode()

// EncodeCBOR encodes any given value to CBOR format.
// It accepts any type that can be marshaled to CBOR.
// Encoding uses RFC 8949/7049bis core deterministic rules.
//
// Parameters:
//   - v: The value to encode (can be any type)
//
// Returns:
//   - CBORData: The CBOR-encoded data as bytes
//   - error: An error if encoding fails, nil otherwise
func EncodeCBOR(v any) (CBORData, error) {
	b, err := deterministicCBORMarshal.Marshal(v)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// EncodeCBORToWriter encodes a value using deterministic CBOR directly into the
// provided writer, avoiding an intermediate byte slice.
func EncodeCBORToWriter(w io.Writer, v any) error {
	return deterministicCBORMarshal.NewEncoder(w).Encode(v)
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

// NewHasher returns a streaming hasher for the requested hash algorithm.
func NewHasher(hashAlgo HashAlgo) hash.Hash {
	switch hashAlgo {
	case SHA384:
		return sha3.New384()
	case SHA512:
		return sha512.New()
	default:
		return sha256.New()
	}
}

type DataHasher struct {
	hasher  hash.Hash
	scratch []byte
}

func NewDataHasher(hashAlgo HashAlgo) *DataHasher {
	return &DataHasher{hasher: NewHasher(hashAlgo)}
}

func appendCBORMajorTypeHeader(dst []byte, major byte, value int) []byte {
	switch {
	case value <= 23:
		return append(dst, byte(major<<5)|byte(value))
	case value <= 0xff:
		return append(dst, byte(major<<5)|24, byte(value))
	case value <= 0xffff:
		dst = append(dst, byte(major<<5)|25)
		return binary.BigEndian.AppendUint16(dst, uint16(value))
	case uint64(value) <= 0xffffffff:
		dst = append(dst, byte(major<<5)|26)
		return binary.BigEndian.AppendUint32(dst, uint32(value))
	default:
		dst = append(dst, byte(major<<5)|27)
		return binary.BigEndian.AppendUint64(dst, uint64(value))
	}
}

// SumTo hashes a deterministic CBOR array whose elements are byte strings or nil.
func (h *DataHasher) SumTo(dst []byte, values ...[]byte) Hash {
	h.hasher.Reset()

	scratch := h.scratch[:0]
	scratch = appendCBORMajorTypeHeader(scratch, 4, len(values))
	_, _ = h.hasher.Write(scratch)

	for _, value := range values {
		scratch = scratch[:0]
		if value == nil {
			scratch = append(scratch, 0xf6)
			_, _ = h.hasher.Write(scratch)
			continue
		}

		scratch = appendCBORMajorTypeHeader(scratch, 2, len(value))
		_, _ = h.hasher.Write(scratch)
		_, _ = h.hasher.Write(value)
	}

	h.scratch = scratch
	return h.hasher.Sum(dst[:0])
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

// ConcatDataAndGenerateCombinedHash concatenates multiple byte slices and generates
// a hash of the combined data using the specified hash algorithm.
//
// The function preallocates memory for efficiency by calculating the total length needed,
// then concatenates all byte slices and generates a hash of the result.
//
// Parameters:
//   - hashAlgo: The hash algorithm to use for generating the hash
//   - data: Variadic parameter accepting multiple byte slices to be concatenated
//
// Returns:
//   - Hash: The generated hash of the concatenated data
func ConcatDataAndGenerateCombinedHash(hashAlgo HashAlgo, data ...[]byte) Hash {
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

// GetBit get bit value at position i (starts at 0)
//
// Parameters:
//   - byteArray: Input byte array
//
// Returns:
//   - bool: Bit value at position i
func GetBit(byteArray []byte, i int) bool {
	byteIdx := i / 8
	if byteIdx >= len(byteArray) {
		return false
	}
	bitIdx := 7 - (i % 8)
	return (byteArray[byteIdx] & (1 << bitIdx)) != 0
}
