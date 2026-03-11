package avlhashtree

import (
	"bytes"
	"testing"

	"github.com/NickOvt/go-chain-trees/utils"
)

func TestAVLHashTree_DefaultInsertHashesKey(t *testing.T) {
	tree := NewAVLHashTree(utils.SHA256)
	key := []byte{0x01, 0x02, 0x03, 0x04}
	keyCBOR, err := utils.EncodeCBOR(key)
	if err != nil {
		t.Fatalf("EncodeCBOR returned error: %v", err)
	}
	hashedKey := utils.GenerateHash(tree.HashAlgo, keyCBOR)

	if err := tree.Insert(key, "value"); err != nil {
		t.Fatalf("Insert returned error: %v", err)
	}

	if tree.Root == nil {
		t.Fatalf("expected root node after insert")
	}
	if !bytes.Equal(tree.Root.Key, hashedKey) {
		t.Fatalf("expected hashed key bytes to be stored, got %x want %x", tree.Root.Key, hashedKey)
	}
	if bytes.Equal(tree.Root.Key, key) {
		t.Fatalf("expected root key to differ from the original un-hashed key")
	}

	got, err := tree.Search(hashedKey)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	gotString, ok := got.(string)
	if !ok || gotString != "value" {
		t.Fatalf("unexpected search result: %#v", got)
	}

	if _, err := tree.Search(key); err == nil {
		t.Fatalf("expected Search to miss the original un-hashed key")
	}

	proof, err := tree.GenerateInclusionExclusionProof(hashedKey)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if !proof.Found {
		t.Fatalf("expected proof to find inserted key")
	}
	if !bytes.Equal(proof.TargetKey, hashedKey) {
		t.Fatalf("expected proof target key to remain hashed, got %x want %x", proof.TargetKey, hashedKey)
	}

	valid, err := tree.VerifyPublicProof(proof.ToPublicProof())
	if err != nil {
		t.Fatalf("VerifyPublicProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected public proof verification to succeed")
	}

	if err := tree.Delete(hashedKey); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if tree.Root != nil {
		t.Fatalf("expected tree to be empty after deleting only node")
	}
}

func TestAVLHashTree_InsertCBORHashesGivenCBORKey(t *testing.T) {
	tree := NewAVLHashTree(utils.SHA256)
	rawKey := []byte{0xaa, 0xbb, 0xcc}
	keyCBOR, err := utils.EncodeCBOR(rawKey)
	if err != nil {
		t.Fatalf("EncodeCBOR returned error: %v", err)
	}
	hashedKey := utils.GenerateHash(tree.HashAlgo, keyCBOR)

	if err := tree.InsertCBOR(keyCBOR, "value"); err != nil {
		t.Fatalf("InsertCBOR returned error: %v", err)
	}

	if tree.Root == nil {
		t.Fatalf("expected root node after InsertCBOR")
	}
	if !bytes.Equal(tree.Root.Key, hashedKey) {
		t.Fatalf("expected hash of provided CBOR key to be stored, got %x want %x", tree.Root.Key, hashedKey)
	}

	got, err := tree.Search(hashedKey)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	gotString, ok := got.(string)
	if !ok || gotString != "value" {
		t.Fatalf("unexpected Search result: %#v", got)
	}

	proof, err := tree.GenerateInclusionExclusionProof(hashedKey)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if !proof.Found {
		t.Fatalf("expected proof to find inserted key")
	}
	if !bytes.Equal(proof.TargetKey, hashedKey) {
		t.Fatalf("expected proof target key to preserve hashed key")
	}

	valid, err := tree.VerifyPublicProof(proof.ToPublicProof())
	if err != nil {
		t.Fatalf("VerifyPublicProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected public proof verification to succeed")
	}

	if err := tree.Delete(hashedKey); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if tree.Root != nil {
		t.Fatalf("expected tree to be empty after Delete")
	}
}

func TestAVLHashTree_InsertHashedStoresGivenHash(t *testing.T) {
	tree := NewAVLHashTree(utils.SHA256)
	key := utils.GenerateHash(tree.HashAlgo, []byte("already-hashed"))

	if err := tree.InsertHashed(key, "value"); err != nil {
		t.Fatalf("InsertHashed returned error: %v", err)
	}

	if tree.Root == nil {
		t.Fatalf("expected root node after InsertHashed")
	}
	if !bytes.Equal(tree.Root.Key, key) {
		t.Fatalf("expected provided hashed key to be stored, got %x want %x", tree.Root.Key, key)
	}

	got, err := tree.Search(key)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	gotString, ok := got.(string)
	if !ok || gotString != "value" {
		t.Fatalf("unexpected Search result: %#v", got)
	}

	proof, err := tree.GenerateInclusionExclusionProof(key)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if !proof.Found {
		t.Fatalf("expected proof to find inserted key")
	}
	if !bytes.Equal(proof.TargetKey, key) {
		t.Fatalf("expected proof target key to preserve hashed key")
	}

	valid, err := tree.VerifyPublicProof(proof.ToPublicProof())
	if err != nil {
		t.Fatalf("VerifyPublicProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected public proof verification to succeed")
	}

	if err := tree.Delete(key); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if tree.Root != nil {
		t.Fatalf("expected tree to be empty after Delete")
	}
}
