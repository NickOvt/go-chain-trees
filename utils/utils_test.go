package utils

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"testing"
)

type sampleStruct struct {
	Name  string
	Count int
}

func TestEncodeCBORAndDecodeCBOR(t *testing.T) {
	original := sampleStruct{Name: "alpha", Count: 3}

	encoded, err := EncodeCBOR(original)
	if err != nil {
		t.Fatalf("EncodeCBOR error: %v", err)
	}

	decoded, err := DecodeCBOR[sampleStruct](encoded)
	if err != nil {
		t.Fatalf("DecodeCBOR error: %v", err)
	}

	if decoded != original {
		t.Fatalf("decoded value mismatch: got %+v want %+v", decoded, original)
	}
}

func TestEncodeCBORError(t *testing.T) {
	_, err := EncodeCBOR(func() {})
	if err == nil {
		t.Fatalf("expected EncodeCBOR error, got nil")
	}
}

func TestDecodeCBORError(t *testing.T) {
	_, err := DecodeCBOR[sampleStruct](CBORData{0xff})
	if err == nil {
		t.Fatalf("expected DecodeCBOR error, got nil")
	}
}

func TestGenerateHashSha256(t *testing.T) {
	data := []byte("hello")
	expected := sha256.Sum256(data)

	got := GenerateHashSha256(data)
	if !bytes.Equal(got, expected[:]) {
		t.Fatalf("sha256 mismatch: got %x want %x", got, expected)
	}
}

func TestGenerateHashSha384(t *testing.T) {
	data := []byte("hello")
	expected := sha3.Sum384(data)

	got := GenerateHashSha384(data)
	if !bytes.Equal(got, expected[:]) {
		t.Fatalf("sha384 mismatch: got %x want %x", got, expected)
	}
}

func TestGenerateHashSha512(t *testing.T) {
	data := []byte("hello")
	expected := sha512.Sum512(data)

	got := GenerateHashSha512(data)
	if !bytes.Equal(got, expected[:]) {
		t.Fatalf("sha512 mismatch: got %x want %x", got, expected)
	}
}

func TestGenerateHash(t *testing.T) {
	data := []byte("hello")
	expected := sha256.Sum256(data)

	got := GenerateHash("unknown", data)
	if !bytes.Equal(got, expected[:]) {
		t.Fatalf("default hash mismatch: got %x want %x", got, expected)
	}

	got = GenerateHash(SHA512, data)
	expected512 := sha512.Sum512(data)
	if !bytes.Equal(got, expected512[:]) {
		t.Fatalf("sha512 mismatch: got %x want %x", got, expected512)
	}
}

func TestGenerateNullHash(t *testing.T) {
	expected := sha256.Sum256(nil)
	got := GenerateNullHash("unknown")
	if !bytes.Equal(got, expected[:]) {
		t.Fatalf("default null hash mismatch: got %x want %x", got, expected)
	}

	expected384 := sha3.Sum384(nil)
	got = GenerateNullHash(SHA384)
	if !bytes.Equal(got, expected384[:]) {
		t.Fatalf("sha384 null hash mismatch: got %x want %x", got, expected384)
	}
}

func TestGetHashAlgoOutputBitCount(t *testing.T) {
	if got := GetHashAlgoOutputBitCount(SHA256); got != 256 {
		t.Fatalf("sha256 bit count mismatch: got %d want 256", got)
	}

	if got := GetHashAlgoOutputBitCount(SHA384); got != 384 {
		t.Fatalf("sha384 bit count mismatch: got %d want 384", got)
	}

	if got := GetHashAlgoOutputBitCount(SHA512); got != 512 {
		t.Fatalf("sha512 bit count mismatch: got %d want 512", got)
	}

	if got := GetHashAlgoOutputBitCount("unknown"); got != 256 {
		t.Fatalf("default bit count mismatch: got %d want 256", got)
	}
}

func TestConcatDataAndGenerateHash(t *testing.T) {
	data1 := CBORData("a")
	data2 := CBORData("bc")

	expected := GenerateHash(SHA256, []byte("abc"))
	got := ConcatDataAndGenerateHash(SHA256, data1, data2)

	if !bytes.Equal(got, expected) {
		t.Fatalf("concat data hash mismatch: got %x want %x", got, expected)
	}
}

func TestConcatHashesAndGenerateHash(t *testing.T) {
	hash1 := GenerateHashSha256([]byte("one"))
	hash2 := GenerateHashSha256([]byte("two"))
	combined := append(append([]byte{}, hash1...), hash2...)

	expected := GenerateHash(SHA256, combined)
	got := ConcatHashesAndGenerateHash(SHA256, hash1, hash2)

	if !bytes.Equal(got, expected) {
		t.Fatalf("concat hash mismatch: got %x want %x", got, expected)
	}
}

func TestEncodeCBORList(t *testing.T) {
	values := []any{"alpha", 42}
	encodedList, err := EncodeCBORList(values...)
	if err != nil {
		t.Fatalf("EncodeCBORList error: %v", err)
	}

	if len(encodedList) != len(values) {
		t.Fatalf("encoded list length mismatch: got %d want %d", len(encodedList), len(values))
	}

	decoded0, err := DecodeCBOR[string](encodedList[0])
	if err != nil {
		t.Fatalf("DecodeCBOR error: %v", err)
	}
	if decoded0 != "alpha" {
		t.Fatalf("decoded value mismatch: got %q want %q", decoded0, "alpha")
	}

	decoded1, err := DecodeCBOR[int](encodedList[1])
	if err != nil {
		t.Fatalf("DecodeCBOR error: %v", err)
	}
	if decoded1 != 42 {
		t.Fatalf("decoded value mismatch: got %d want %d", decoded1, 42)
	}
}

func TestEncodeCBORListError(t *testing.T) {
	_, err := EncodeCBORList("ok", func() {})
	if err == nil {
		t.Fatalf("expected EncodeCBORList error, got nil")
	}
}
