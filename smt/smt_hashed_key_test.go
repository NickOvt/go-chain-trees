package smt

import (
	"bytes"
	"testing"

	"github.com/NickOvt/go-chain-trees/utils"
)

func TestSMT_InsertHashedStoresGivenHash(t *testing.T) {
	tree := NewSMT(utils.SHA256, true)
	key := utils.GenerateHash(tree.HashAlgo, []byte("already-hashed"))
	data := []byte("value")

	inserted, err := tree.InsertHashed(key, data)
	if err != nil {
		t.Fatalf("InsertHashed returned error: %v", err)
	}
	if !inserted {
		t.Fatalf("expected InsertHashed to report successful insertion")
	}

	leaf := findLeafByStoredKey(tree.GetRoot(), key)
	if leaf == nil {
		t.Fatalf("expected to find inserted leaf by stored hashed key")
	}
	if !bytes.Equal(leaf.Key, key) {
		t.Fatalf("expected stored key to match provided hash, got %x want %x", leaf.Key, key)
	}
	if !bytes.Equal(leaf.Data, data) {
		t.Fatalf("expected stored leaf data to match insert payload")
	}

	proof, err := tree.GenerateInclusionExclusionProof(key)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if !bytes.Equal(proof.Key, key) {
		t.Fatalf("expected proof key to preserve hashed key")
	}

	valid, err := tree.VerifyProof(proof)
	if err != nil {
		t.Fatalf("VerifyProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected hashed-key proof verification to succeed")
	}
}

func findLeafByStoredKey(node *Node, key utils.Hash) *Node {
	if node == nil {
		return nil
	}

	if node.IsLeaf {
		if bytes.Equal(node.Key, key) {
			return node
		}
		return nil
	}

	if found := findLeafByStoredKey(node.LeftNode, key); found != nil {
		return found
	}

	return findLeafByStoredKey(node.RightNode, key)
}
