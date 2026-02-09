package utils

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"math/bits"

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

// ReverseBits returns a new byte array with both the order of bytes and the bits
// within each byte reversed. For example, if the input is [0b10110000, 0b00001111],
// the output will be [0b11110000, 0b00001101].
//
// The function does not modify the original byte array.
func ReverseBits(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = bits.Reverse8(b)
	}
	return result
}

// FindCommonBitPrefix returns the longest common bit prefix of two byte arrays.
// The function compares the byte arrays bit by bit and returns a new byte array
// containing all bits that match from the beginning until the first differing bit.
//
// If the common prefix doesn't end on a byte boundary, the last byte contains only
// the matching bits with the remaining bits zeroed out.
//
// If the arrays have different lengths, comparison stops at the end of the shorter array.
// If there is no common prefix, an empty byte array is returned with prefixLen = 0.
//
// Parameters:
//   - a: First byte array
//   - b: Second byte array
//
// Returns:
//   - []byte: A new byte array containing the common bit prefix
//   - int: Number of bits in the common prefix
//
// Example:
//   - a = [0b11110000]
//   - b = [0b11111111]
//   - result = [0b11110000], prefixLen = 4, paddingBits = 4
func FindCommonBitPrefix(a, b []byte) ([]byte, int, int) {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	prefixBits := 0

	// Find how many bits match
	for i := 0; i < minLen; i++ {
		xor := a[i] ^ b[i]
		if xor == 0 {
			// All 8 bits in this byte match
			prefixBits += 8
		} else {
			// Find first differing bit using leading zeros count
			prefixBits += bits.LeadingZeros8(xor)
			break
		}
	}

	// Calculate how many bytes we need
	prefixBytes := prefixBits / 8
	remainingBits := prefixBits % 8

	result := make([]byte, prefixBytes)
	copy(result, a[:prefixBytes])

	paddingBits := 0
	// If there are remaining bits, add a partial byte
	if remainingBits > 0 {
		paddingBits = 8 - remainingBits
		mask := byte(0xFF << paddingBits)
		result = append(result, a[prefixBytes]&mask)
	}

	return result, prefixBits, paddingBits
}

// RemoveFirstNBits removes the first n bits from a byte array and returns the remaining bits.
// The result is left-aligned (shifted to the beginning) and padded with zeros at the end
// if necessary to fill the last byte.
//
// Parameters:
//   - data: The input byte array
//   - n: Number of bits to remove from the beginning
//
// Returns:
//   - []byte: A new byte array with the first n bits removed, padded with zeros at the end
//   - int: Number of bits in the result (excluding padding)
//
// Example:
//   - data = [0b11110000, 0b10101010]
//   - n = 4
//   - result = [0b00001010, 0b10100000], resultBits = 12
//   - (removed first 4 bits, shifted remaining bits left, 12 bits remain)
func RemoveFirstNBits(data []byte, n int) ([]byte, int, int) {
	if n <= 0 {
		return append([]byte(nil), data...), len(data) * 8, 0
	}

	totalBits := len(data) * 8
	if n >= totalBits {
		return []byte{}, 0, 0
	}

	remainingBits := totalBits - n
	resultBytes := (remainingBits + 7) / 8 // Ceiling division
	paddingBits := (resultBytes * 8) - remainingBits

	result := make([]byte, resultBytes)

	skipBytes := n / 8
	skipBits := n % 8

	for i := 0; i < resultBytes; i++ {
		srcByteIdx := skipBytes + i

		if srcByteIdx < len(data) {
			// Take remaining bits from current byte
			result[i] = data[srcByteIdx] << skipBits

			// Take leading bits from next byte if needed
			if skipBits > 0 && srcByteIdx+1 < len(data) {
				result[i] |= data[srcByteIdx+1] >> (8 - skipBits)
			}
		}
	}

	return result, remainingBits, paddingBits
}

// FindCommonBitPrefixWithLen returns the longest common bit prefix of two byte arrays,
// considering only the specified number of meaningful bits in each array.
//
// This is essential when comparing bit sequences that don't fill complete bytes,
// as it prevents padding bits from affecting the comparison.
//
// Parameters:
//   - a: First byte array
//   - aBitLen: Number of meaningful bits in 'a'
//   - b: Second byte array
//   - bBitLen: Number of meaningful bits in 'b'
//
// Returns:
//   - []byte: A new byte array containing the common bit prefix
//   - int: Number of bits in the common prefix
//   - int: Number of padding bits in the last byte of the result
func FindCommonBitPrefixWithLen(a []byte, aBitLen int, b []byte, bBitLen int) ([]byte, int, int) {
	// Only compare up to the shorter of the two meaningful lengths
	maxBitsToCompare := aBitLen
	if bBitLen < maxBitsToCompare {
		maxBitsToCompare = bBitLen
	}

	if maxBitsToCompare <= 0 {
		return []byte{}, 0, 0
	}

	prefixBits := 0

	// Compare byte by byte, but respect the bit length limits
	maxBytes := (maxBitsToCompare + 7) / 8

	for i := 0; i < maxBytes; i++ {
		if i >= len(a) || i >= len(b) {
			break
		}

		// Calculate how many bits to compare in this byte
		bitsInThisByte := 8
		if (i+1)*8 > maxBitsToCompare {
			bitsInThisByte = maxBitsToCompare - i*8
		}

		xor := a[i] ^ b[i]

		if xor == 0 && bitsInThisByte == 8 {
			// All 8 bits in this byte match
			prefixBits += 8
		} else {
			// Find first differing bit
			matchingBits := bits.LeadingZeros8(xor)

			// Don't count matches beyond what we should compare
			if matchingBits > bitsInThisByte {
				matchingBits = bitsInThisByte
			}

			prefixBits += matchingBits
			break
		}
	}

	// Calculate how many bytes we need
	prefixBytes := prefixBits / 8
	remainingBits := prefixBits % 8

	result := make([]byte, prefixBytes)
	copy(result, a[:prefixBytes])

	paddingBits := 0
	// If there are remaining bits, add a partial byte
	if remainingBits > 0 {
		paddingBits = 8 - remainingBits
		mask := byte(0xFF << paddingBits)
		result = append(result, a[prefixBytes]&mask)
	}

	return result, prefixBits, paddingBits
}
