package test

import (
	"testing"

	"github.com/NickOvt/go-chain-trees/smt"
	"github.com/NickOvt/go-chain-trees/utils"
)

func TestSMT_RightSubtree_OneInsertedNode(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "1", used)

	insertKeys(t, tree, k1)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.LeftNode != nil {
		t.Fatalf("expected root.LeftNode to be nil")
	}
	if root.RightNode == nil || !root.RightNode.IsLeaf {
		t.Fatalf("expected one leaf directly under root.RightNode")
	}

	assertChildSideBit(t, root.RightNode, true)
	assertSubtreeIntegrity(t, root.RightNode)
	assertLeafHashesMatch(t, root.RightNode, [][]byte{k1})
}

func TestSMT_RightSubtree_TwoInsertedNodesSamePath(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "1110", used)
	k2 := findKeyWithHashPrefix(t, "1111", used)

	insertKeys(t, tree, k1, k2)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.LeftNode != nil {
		t.Fatalf("expected root.LeftNode to stay nil")
	}

	right := root.RightNode
	if right == nil || right.IsLeaf {
		t.Fatalf("expected root.RightNode to be a branch after two inserts")
	}

	_, pathBits := smt.CalculateKeyFromPath(right.Path)
	if pathBits < 2 {
		t.Fatalf("expected shared right path length >= 2 bits, got %d", pathBits)
	}
	assertNodePathBits(t, right, "111")

	if right.LeftNode == nil || right.RightNode == nil || !right.LeftNode.IsLeaf || !right.RightNode.IsLeaf {
		t.Fatalf("expected root.RightNode to have two leaf children")
	}

	assertChildSideBit(t, right, true)
	assertSubtreeIntegrity(t, right)
	assertLeafHashesMatch(t, right, [][]byte{k1, k2})
	tree.PrintTree()
}

func TestSMT_RightSubtree_ThreeInsertedNodes(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "1110", used)
	k2 := findKeyWithHashPrefix(t, "1111", used)
	k3 := findKeyWithHashPrefix(t, "1101", used)

	insertKeys(t, tree, k1, k2, k3)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.LeftNode != nil {
		t.Fatalf("expected root.LeftNode to stay nil")
	}

	right := root.RightNode
	if right == nil || right.IsLeaf {
		t.Fatalf("expected root.RightNode to remain a branch")
	}

	if right.LeftNode == nil || right.RightNode == nil {
		t.Fatalf("expected root.RightNode to have both children")
	}
	if !right.LeftNode.IsLeaf || right.RightNode.IsLeaf {
		t.Fatalf("expected left child leaf and right child branch after 3 inserts")
	}
	assertNodePathBits(t, right, "11")
	assertNodePathBits(t, right.RightNode, "1")

	assertSubtreeIntegrity(t, right)
	if countLeaves(right) != 3 {
		t.Fatalf("expected 3 leaves in root right subtree, got %d", countLeaves(right))
	}
	assertLeafHashesMatch(t, right, [][]byte{k1, k2, k3})
	tree.PrintTree()
}

func TestSMT_RightSubtree_FourInsertedNodes(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "1100", used)
	k2 := findKeyWithHashPrefix(t, "1101", used)
	k3 := findKeyWithHashPrefix(t, "1110", used)
	k4 := findKeyWithHashPrefix(t, "1111", used)

	insertKeys(t, tree, k1, k2, k3, k4)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.LeftNode != nil {
		t.Fatalf("expected root.LeftNode to stay nil")
	}

	right := root.RightNode
	if right == nil || right.IsLeaf {
		t.Fatalf("expected root.RightNode to be a branch")
	}

	if right.LeftNode == nil || right.RightNode == nil {
		t.Fatalf("expected root.RightNode to have both children")
	}
	if right.LeftNode.IsLeaf || right.RightNode.IsLeaf {
		t.Fatalf("expected both root.RightNode children to be branches after 4 inserts")
	}
	assertNodePathBits(t, right, "11")
	assertNodePathBits(t, right.LeftNode, "0")
	assertNodePathBits(t, right.RightNode, "1")

	for _, b := range []*smt.Node{right.LeftNode, right.RightNode} {
		if b.LeftNode == nil || b.RightNode == nil || !b.LeftNode.IsLeaf || !b.RightNode.IsLeaf {
			t.Fatalf("expected each second-level branch to contain exactly two leaves")
		}
	}

	assertSubtreeIntegrity(t, right)
	if countLeaves(right) != 4 {
		t.Fatalf("expected 4 leaves in root right subtree, got %d", countLeaves(right))
	}
	assertLeafHashesMatch(t, right, [][]byte{k1, k2, k3, k4})
	tree.PrintTree()
}

func TestSMT_RightSubtree_FiveInsertedNodes(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "1111", used)
	k2 := findKeyWithHashPrefix(t, "1110", used)
	k3 := findKeyWithHashPrefix(t, "1101", used)
	k4 := findKeyWithHashPrefix(t, "1100", used)
	k5 := findKeyWithHashPrefix(t, "1011", used)

	insertKeys(t, tree, k1, k2, k3, k4, k5)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.LeftNode != nil {
		t.Fatalf("expected root.LeftNode to stay nil")
	}

	right := root.RightNode
	if right == nil || right.IsLeaf {
		t.Fatalf("expected root.RightNode to be a branch")
	}

	if right.LeftNode == nil || right.RightNode == nil {
		t.Fatalf("expected root.RightNode to have both children")
	}
	if !right.LeftNode.IsLeaf || right.RightNode.IsLeaf {
		t.Fatalf("expected root.right.left to be leaf and root.right.right to be branch")
	}
	assertNodePathBits(t, right, "1")
	assertNodePathBits(t, right.RightNode, "1")
	assertNodePathBits(t, right.RightNode.RightNode, "1")
	assertNodePathBits(t, right.RightNode.LeftNode, "0")

	assertSubtreeIntegrity(t, right)
	if countLeaves(right) != 5 {
		t.Fatalf("expected 5 leaves in root right subtree, got %d", countLeaves(right))
	}
	assertLeafHashesMatch(t, right, [][]byte{k1, k2, k3, k4, k5})
	tree.PrintTree()
}

func TestSMT_RightSubtree_TwoInsertedNodes_LongCommonPrefixes(t *testing.T) {
	testCases := []struct {
		name             string
		commonPrefixBits string
	}{
		{name: "7_bits", commonPrefixBits: "1100110"},
		{name: "11_bits", commonPrefixBits: "10100111011"},
		{name: "12_bits", commonPrefixBits: "101001110110"},
		{name: "15_bits", commonPrefixBits: "101001110110100"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tree := smt.NewSMT(utils.SHA256)
			used := map[string]struct{}{}

			k1 := findKeyWithHashPrefix(t, tc.commonPrefixBits+"0", used)
			k2 := findKeyWithHashPrefix(t, tc.commonPrefixBits+"1", used)

			insertKeys(t, tree, k1, k2)

			root := tree.GetRoot()
			assertRootIntegrity(t, root)
			if root.LeftNode != nil {
				t.Fatalf("expected root.LeftNode to stay nil")
			}

			right := root.RightNode
			if right == nil || right.IsLeaf {
				t.Fatalf("expected root.RightNode to be a branch")
			}
			assertNodePathBits(t, right, tc.commonPrefixBits)

			if right.LeftNode == nil || right.RightNode == nil || !right.LeftNode.IsLeaf || !right.RightNode.IsLeaf {
				t.Fatalf("expected root.RightNode to have two leaf children")
			}

			assertSubtreeIntegrity(t, right)
			assertLeafHashesMatch(t, right, [][]byte{k1, k2})
			tree.PrintTree()
		})
	}
}
