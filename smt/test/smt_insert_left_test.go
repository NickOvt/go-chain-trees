package test

import (
	"fmt"
	"testing"

	"github.com/NickOvt/go-chain-trees/smt"
	"github.com/NickOvt/go-chain-trees/utils"
)

func TestSMT_LeftSubtree_OneInsertedNode(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "0", used)

	insertKeys(t, tree, k1)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.RightNode != nil {
		t.Fatalf("expected root.RightNode to be nil")
	}
	if root.LeftNode == nil || !root.LeftNode.IsLeaf {
		t.Fatalf("expected one leaf directly under root.LeftNode")
	}

	assertChildSideBit(t, root.LeftNode, false)
	assertSubtreeIntegrity(t, root.LeftNode)
	assertLeafHashesMatch(t, root.LeftNode, [][]byte{k1})
}

func TestSMT_LeftSubtree_TwoInsertedNodesSamePath(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "0000", used)
	k2 := findKeyWithHashPrefix(t, "0001", used)

	insertKeys(t, tree, k1, k2)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.RightNode != nil {
		t.Fatalf("expected root.RightNode to stay nil")
	}

	left := root.LeftNode
	if left == nil || left.IsLeaf {
		t.Fatalf("expected root.LeftNode to be a branch after two inserts")
	}

	_, pathBits := smt.CalculateKeyFromPath(left.Path)
	if pathBits < 2 {
		t.Fatalf("expected shared left path length >= 2 bits, got %d", pathBits)
	}

	if left.LeftNode == nil || left.RightNode == nil || !left.LeftNode.IsLeaf || !left.RightNode.IsLeaf {
		t.Fatalf("expected root.LeftNode to have two leaf children")
	}

	assertChildSideBit(t, left, false)
	assertSubtreeIntegrity(t, left)
	assertLeafHashesMatch(t, left, [][]byte{k1, k2})
}

func TestSMT_LeftSubtree_ThreeInsertedNodes(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "0000", used)
	k2 := findKeyWithHashPrefix(t, "0001", used)
	k3 := findKeyWithHashPrefix(t, "0010", used)

	insertKeys(t, tree, k1, k2, k3)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.RightNode != nil {
		t.Fatalf("expected root.RightNode to stay nil")
	}

	left := root.LeftNode
	if left == nil || left.IsLeaf {
		t.Fatalf("expected root.LeftNode to remain a branch")
	}

	if left.LeftNode == nil || left.RightNode == nil {
		t.Fatalf("expected root.LeftNode to have both children")
	}
	if left.LeftNode.IsLeaf || !left.RightNode.IsLeaf {
		t.Fatalf("expected left child branch and right child leaf after 3 inserts")
	}

	assertSubtreeIntegrity(t, left)
	if countLeaves(left) != 3 {
		t.Fatalf("expected 3 leaves in root left subtree, got %d", countLeaves(left))
	}
	assertLeafHashesMatch(t, left, [][]byte{k1, k2, k3})
}

func TestSMT_LeftSubtree_FourInsertedNodes(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256)
	used := map[string]struct{}{}
	k1 := findKeyWithHashPrefix(t, "0000", used)
	k2 := findKeyWithHashPrefix(t, "0001", used)
	k3 := findKeyWithHashPrefix(t, "0010", used)
	k4 := findKeyWithHashPrefix(t, "0011", used)

	insertKeys(t, tree, k1, k2, k3, k4)

	root := tree.GetRoot()
	assertRootIntegrity(t, root)
	if root.RightNode != nil {
		t.Fatalf("expected root.RightNode to stay nil")
	}

	left := root.LeftNode
	if left == nil || left.IsLeaf {
		t.Fatalf("expected root.LeftNode to be a branch")
	}

	if left.LeftNode == nil || left.RightNode == nil {
		t.Fatalf("expected root.LeftNode to have both children")
	}
	if left.LeftNode.IsLeaf || left.RightNode.IsLeaf {
		t.Fatalf("expected both root.LeftNode children to be branches after 4 inserts")
	}

	for _, b := range []*smt.Node{left.LeftNode, left.RightNode} {
		if b.LeftNode == nil || b.RightNode == nil || !b.LeftNode.IsLeaf || !b.RightNode.IsLeaf {
			t.Fatalf("expected each second-level branch to contain exactly two leaves")
		}
	}

	assertSubtreeIntegrity(t, left)
	if countLeaves(left) != 4 {
		t.Fatalf("expected 4 leaves in root left subtree, got %d", countLeaves(left))
	}
	assertLeafHashesMatch(t, left, [][]byte{k1, k2, k3, k4})
}

func insertKeys(t *testing.T, tree *smt.SMT, keys ...[]byte) {
	t.Helper()
	for i, key := range keys {
		ok, err := tree.Insert(key, []byte(fmt.Sprintf("value-%d", i)))
		if err != nil {
			t.Fatalf("insert failed for key %q: %v", key, err)
		}
		if !ok {
			t.Fatalf("insert returned ok=false for key %q", key)
		}
	}
}

func assertRootIntegrity(t *testing.T, root *smt.Node) {
	t.Helper()
	if root == nil {
		t.Fatalf("root is nil")
	}
	if root.IsLeaf {
		t.Fatalf("root must be a branch")
	}
	if len(root.Hash) == 0 {
		t.Fatalf("root hash is empty")
	}
}

func assertSubtreeIntegrity(t *testing.T, node *smt.Node) {
	t.Helper()
	if node == nil {
		t.Fatalf("node is nil")
	}
	if len(node.Hash) == 0 {
		t.Fatalf("node hash is empty")
	}

	if node.IsLeaf {
		if node.LeftNode != nil || node.RightNode != nil {
			t.Fatalf("leaf node must not have children")
		}
		if len(node.Key) == 0 {
			t.Fatalf("leaf key is empty")
		}
		if len(node.Path) == 0 {
			t.Fatalf("leaf path is empty")
		}
		_, bits := smt.CalculateKeyFromPath(node.Path)
		if bits <= 0 {
			t.Fatalf("leaf path has no meaningful bits")
		}
		return
	}

	if node.LeftNode == nil || node.RightNode == nil {
		t.Fatalf("branch node must have both children")
	}

	assertChildSideBit(t, node.LeftNode, false)
	assertChildSideBit(t, node.RightNode, true)
	assertSubtreeIntegrity(t, node.LeftNode)
	assertSubtreeIntegrity(t, node.RightNode)
}

func assertChildSideBit(t *testing.T, child *smt.Node, wantRight bool) {
	t.Helper()
	if child == nil {
		t.Fatalf("child is nil")
	}

	keyBits, bitLen := smt.CalculateKeyFromPath(child.Path)
	if bitLen <= 0 {
		t.Fatalf("child path has no meaningful bits")
	}

	gotRight := utils.GetBit(keyBits, 0)
	if gotRight != wantRight {
		t.Fatalf("child is on wrong side: gotRight=%v, wantRight=%v", gotRight, wantRight)
	}
}

func countLeaves(node *smt.Node) int {
	if node == nil {
		return 0
	}
	if node.IsLeaf {
		return 1
	}

	return countLeaves(node.LeftNode) + countLeaves(node.RightNode)
}

func assertLeafHashesMatch(t *testing.T, root *smt.Node, insertedKeys [][]byte) {
	t.Helper()
	leaves := collectLeaves(root)

	if len(leaves) != len(insertedKeys) {
		t.Fatalf("leaf count mismatch: got %d, want %d", len(leaves), len(insertedKeys))
	}

	expected := make(map[string]struct{}, len(insertedKeys))
	for _, key := range insertedKeys {
		hash := utils.GenerateHash(utils.SHA256, key)
		expected[string(hash)] = struct{}{}
	}

	for _, leaf := range leaves {
		if _, ok := expected[string(leaf.Key)]; !ok {
			t.Fatalf("unexpected leaf key hash found in tree")
		}
	}
}

func collectLeaves(node *smt.Node) []*smt.Node {
	if node == nil {
		return nil
	}
	if node.IsLeaf {
		return []*smt.Node{node}
	}

	result := collectLeaves(node.LeftNode)
	result = append(result, collectLeaves(node.RightNode)...)
	return result
}

func findKeyWithHashPrefix(t *testing.T, prefix string, used map[string]struct{}) []byte {
	t.Helper()
	for i := 0; i < 2_000_000; i++ {
		key := []byte(fmt.Sprintf("k-%s-%d", prefix, i))
		if _, ok := used[string(key)]; ok {
			continue
		}

		hash := utils.GenerateHash(utils.SHA256, key)
		if hasBitPrefix(hash, prefix) {
			used[string(key)] = struct{}{}
			return key
		}
	}

	t.Fatalf("failed to find key with hash prefix %q", prefix)
	return nil
}

func hasBitPrefix(hash []byte, prefix string) bool {
	for i, ch := range prefix {
		want := ch == '1'
		if utils.GetBit(hash, i) != want {
			return false
		}
	}
	return true
}
