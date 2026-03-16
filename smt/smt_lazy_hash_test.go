package smt

import (
	"testing"

	"github.com/NickOvt/go-chain-trees/utils"
)

func TestSMT_InsertDefersHashAndPathEncodingUntilRead(t *testing.T) {
	tree := NewSMT(utils.SHA256, true)

	ok, err := tree.Insert([]byte("lazy-key"), []byte("lazy-value"))
	if err != nil || !ok {
		t.Fatalf("Insert failed: ok=%v err=%v", ok, err)
	}

	root := tree.Root
	if root == nil {
		t.Fatalf("expected internal root to exist after insert")
	}
	if len(root.Hash) != 0 {
		t.Fatalf("expected root hash to stay invalidated until read")
	}

	var leaf *Node
	if root.LeftNode != nil {
		leaf = root.LeftNode
	} else {
		leaf = root.RightNode
	}
	if leaf == nil {
		t.Fatalf("expected inserted leaf to exist")
	}
	if len(leaf.Hash) != 0 {
		t.Fatalf("expected leaf hash to stay invalidated until read")
	}
	if leaf.encodedPath != nil {
		t.Fatalf("expected encoded path to be generated lazily")
	}

	root = tree.GetRoot()
	if len(root.Hash) == 0 {
		t.Fatalf("expected root hash to materialize on read")
	}
	if len(leaf.Hash) == 0 {
		t.Fatalf("expected leaf hash to materialize on read")
	}
	if leaf.encodedPath == nil {
		t.Fatalf("expected encoded path to materialize on read")
	}
}
