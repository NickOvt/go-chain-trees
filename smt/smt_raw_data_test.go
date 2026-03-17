package smt

import (
	"bytes"
	"testing"

	"github.com/NickOvt/go-chain-trees/utils"
)

func firstLeafFromRoot(root *Node) *Node {
	if root == nil {
		return nil
	}
	if root.LeftNode != nil {
		return root.LeftNode
	}
	return root.RightNode
}

func TestSMT_InsertStoresRawLeafDataAndHashesCBORArray(t *testing.T) {
	tree := NewSMT(utils.SHA256, false)
	payload := []byte("raw-payload")

	ok, err := tree.Insert([]byte("key"), payload)
	if err != nil || !ok {
		t.Fatalf("Insert failed: ok=%v err=%v", ok, err)
	}

	leaf := firstLeafFromRoot(tree.GetRoot())
	if leaf == nil || !leaf.IsLeaf {
		t.Fatalf("expected inserted leaf")
	}
	if !bytes.Equal(leaf.Data, payload) {
		t.Fatalf("expected raw payload to be stored in leaf")
	}

	encodedTuple, err := utils.EncodeCBOR([]any{leaf.Path.Encode(), payload})
	if err != nil {
		t.Fatalf("failed to encode expected tuple: %v", err)
	}
	expectedHash := utils.GenerateHash(tree.HashAlgo, encodedTuple)
	if !bytes.Equal(leaf.Hash, expectedHash) {
		t.Fatalf("unexpected leaf hash")
	}
}
