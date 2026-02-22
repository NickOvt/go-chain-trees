package test

import (
	"bytes"
	"testing"

	"github.com/NickOvt/go-chain-trees/smt"
	"github.com/NickOvt/go-chain-trees/utils"
)

func TestSMT_GenerateInclusionExclusionProof_Inclusion_LeafToRootOrderAndHashChain(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	k00 := findKeyWithHashPrefix(t, "00", used)
	k01 := findKeyWithHashPrefix(t, "01", used)
	k10 := findKeyWithHashPrefix(t, "10", used)
	k11 := findKeyWithHashPrefix(t, "11", used)
	insertKeys(t, tree, k00, k01, k10, k11)

	targetHash := utils.GenerateHash(tree.HashAlgo, k00)
	proof, err := tree.GenerateInclusionExclusionProof(targetHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if proof == nil {
		t.Fatalf("expected proof, got nil")
	}
	if len(proof.Path) < 2 {
		t.Fatalf("expected proof path with at least leaf + root witness, got %d", len(proof.Path))
	}
	if !bytes.Equal(proof.Root, tree.GetRoot().Hash) {
		t.Fatalf("proof root hash mismatch")
	}

	first := proof.Path[0]
	if len(first.Data) == 0 {
		t.Fatalf("expected first proof node to contain leaf data")
	}
	if len(first.Hash) != 0 {
		t.Fatalf("expected first proof node hash witness to be empty for inclusion proof")
	}

	leaf := findLeafByKeyHash(tree.GetRoot(), targetHash)
	if leaf == nil {
		t.Fatalf("failed to find target leaf in tree")
	}
	if !bytes.Equal(first.Path, leaf.Path.Encode()) {
		t.Fatalf("first proof node path does not match target leaf path")
	}
	if !bytes.Equal(first.Data, leaf.Data) {
		t.Fatalf("first proof node data does not match target leaf data")
	}

	expectedLeafHash := utils.ConcatDataAndGenerateCombinedHash(tree.HashAlgo, first.Path, first.Data)
	if !bytes.Equal(expectedLeafHash, leaf.Hash) {
		t.Fatalf("first proof node does not recreate target leaf hash")
	}

	for i := 1; i < len(proof.Path); i++ {
		if len(proof.Path[i].Data) != 0 {
			t.Fatalf("expected proof node %d to contain sibling hash witness, not leaf data", i)
		}
	}

	last := proof.Path[len(proof.Path)-1]
	if len(last.Path) != 0 {
		t.Fatalf("expected last proof node to correspond to root path")
	}

	recomputedRoot := recomputeRootFromProofBySpec(t, tree.HashAlgo, proof)
	if !bytes.Equal(recomputedRoot, proof.Root) {
		t.Fatalf("spec recomputed root mismatch with proof root")
	}
	if !bytes.Equal(recomputedRoot, tree.GetRoot().Hash) {
		t.Fatalf("spec recomputed root mismatch with tree root")
	}
}

func TestSMT_GenerateInclusionExclusionProof_Exclusion_MissingRootChild(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	existingKey := findKeyWithHashPrefix(t, "0", used)
	insertKeys(t, tree, existingKey)

	missingKey := findKeyWithHashPrefix(t, "1", used)
	missingKeyHash := utils.GenerateHash(tree.HashAlgo, missingKey)

	proof, err := tree.GenerateInclusionExclusionProof(missingKeyHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}
	if proof == nil {
		t.Fatalf("expected proof, got nil")
	}
	if len(proof.Path) != 1 {
		t.Fatalf("expected one-step exclusion proof for missing root child, got %d", len(proof.Path))
	}
	if !bytes.Equal(proof.Root, tree.GetRoot().Hash) {
		t.Fatalf("proof root hash mismatch")
	}

	step := proof.Path[0]
	if len(step.Path) != 0 {
		t.Fatalf("expected exclusion witness to be rooted at tree root path")
	}
	if len(step.Data) != 0 {
		t.Fatalf("did not expect leaf data in missing-child exclusion witness")
	}
	if len(step.Hash) == 0 {
		t.Fatalf("expected sibling hash witness in exclusion proof")
	}

	existingKeyHash := utils.GenerateHash(tree.HashAlgo, existingKey)
	existingLeaf := findLeafByKeyHash(tree.GetRoot(), existingKeyHash)
	if existingLeaf == nil {
		t.Fatalf("failed to find existing leaf")
	}
	if !bytes.Equal(step.Hash, existingLeaf.Hash) {
		t.Fatalf("exclusion witness hash does not match existing sibling leaf hash")
	}
}

func TestSMT_GenerateInclusionExclusionProof_EmptyTree_ReturnsError(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	target := utils.GenerateHash(tree.HashAlgo, []byte("missing-key"))

	proof, err := tree.GenerateInclusionExclusionProof(target)
	if err == nil {
		t.Fatalf("expected error for empty tree proof generation")
	}
	if proof != nil {
		t.Fatalf("expected nil proof when tree is empty")
	}
}

func TestSMT_InclusionExclusionProof_ToPublicProof_ReturnsRootAndPathTuples(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	k00 := findKeyWithHashPrefix(t, "00", used)
	k01 := findKeyWithHashPrefix(t, "01", used)
	insertKeys(t, tree, k00, k01)

	targetHash := utils.GenerateHash(tree.HashAlgo, k00)
	proof, err := tree.GenerateInclusionExclusionProof(targetHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}

	publicProof := proof.ToPublicProof()
	if publicProof == nil {
		t.Fatalf("expected public proof, got nil")
	}
	if !bytes.Equal(publicProof.Root, proof.Root) {
		t.Fatalf("public proof root mismatch")
	}
	if len(publicProof.Path) != len(proof.Path) {
		t.Fatalf("public proof path length mismatch")
	}
	if len(publicProof.Path) == 0 {
		t.Fatalf("expected non-empty public proof path")
	}

	if !bytes.Equal(publicProof.Path[0][0], proof.Path[0].Path) {
		t.Fatalf("public proof tuple[0] path mismatch")
	}
	if !bytes.Equal(publicProof.Path[0][1], proof.Path[0].Data) {
		t.Fatalf("public proof tuple[0] should contain leaf data")
	}

	for i := 1; i < len(publicProof.Path); i++ {
		if !bytes.Equal(publicProof.Path[i][0], proof.Path[i].Path) {
			t.Fatalf("public proof tuple[%d] path mismatch", i)
		}
		if !bytes.Equal(publicProof.Path[i][1], proof.Path[i].Hash) {
			t.Fatalf("public proof tuple[%d] should contain hash witness", i)
		}
	}
}

func TestSMT_InclusionExclusionProof_ToPublicProof_NilReceiver(t *testing.T) {
	var proof *smt.InclusionExclusionProof
	if proof.ToPublicProof() != nil {
		t.Fatalf("expected nil public proof for nil receiver")
	}
}

func TestSMT_VerifyProof_Inclusion(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	k00 := findKeyWithHashPrefix(t, "00", used)
	k01 := findKeyWithHashPrefix(t, "01", used)
	insertKeys(t, tree, k00, k01)

	targetHash := utils.GenerateHash(tree.HashAlgo, k00)
	proof, err := tree.GenerateInclusionExclusionProof(targetHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}

	valid, err := tree.VerifyProof(proof)
	if err != nil {
		t.Fatalf("VerifyProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected valid inclusion proof")
	}
}

func TestSMT_VerifyPublicProof_Inclusion(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	k00 := findKeyWithHashPrefix(t, "00", used)
	k01 := findKeyWithHashPrefix(t, "01", used)
	insertKeys(t, tree, k00, k01)

	targetHash := utils.GenerateHash(tree.HashAlgo, k00)
	proof, err := tree.GenerateInclusionExclusionProof(targetHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}

	valid, err := tree.VerifyPublicProof(proof.ToPublicProof())
	if err != nil {
		t.Fatalf("VerifyPublicProof returned error: %v", err)
	}
	if !valid {
		t.Fatalf("expected valid public inclusion proof")
	}
}

func TestSMT_VerifyPublicProof_ExclusionMissingRootChild(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	existingKey := findKeyWithHashPrefix(t, "0", used)
	insertKeys(t, tree, existingKey)

	missingKey := findKeyWithHashPrefix(t, "1", used)
	missingKeyHash := utils.GenerateHash(tree.HashAlgo, missingKey)

	proof, err := tree.GenerateInclusionExclusionProof(missingKeyHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}

	valid, err := tree.VerifyPublicProof(proof.ToPublicProof())
	if err == nil {
		t.Fatalf("expected error for exclusion proof represented in simplified public format")
	}
	if valid {
		t.Fatalf("expected simplified public exclusion proof to be invalid")
	}
}

func TestSMT_VerifyPublicProof_TamperedRoot(t *testing.T) {
	tree := smt.NewSMT(utils.SHA256, true)
	used := map[string]struct{}{}

	k00 := findKeyWithHashPrefix(t, "00", used)
	k01 := findKeyWithHashPrefix(t, "01", used)
	insertKeys(t, tree, k00, k01)

	targetHash := utils.GenerateHash(tree.HashAlgo, k00)
	proof, err := tree.GenerateInclusionExclusionProof(targetHash)
	if err != nil {
		t.Fatalf("GenerateInclusionExclusionProof returned error: %v", err)
	}

	publicProof := proof.ToPublicProof()
	publicProof.Root = utils.GenerateHash(tree.HashAlgo, []byte("tampered-root"))

	valid, err := tree.VerifyPublicProof(publicProof)
	if err == nil {
		t.Fatalf("expected error for tampered public proof root")
	}
	if valid {
		t.Fatalf("expected tampered public proof to be invalid")
	}
}

func recomputeRootFromProofBySpec(t *testing.T, hashAlgo utils.HashAlgo, proof *smt.InclusionExclusionProof) []byte {
	t.Helper()

	if proof == nil || len(proof.Path) == 0 {
		t.Fatalf("cannot recompute root from empty proof")
	}
	if len(proof.Path[0].Data) == 0 {
		t.Fatalf("first proof node must contain leaf data for inclusion verification")
	}

	currentHash := utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[0].Path, proof.Path[0].Data)

	for i := 1; i < len(proof.Path); i++ {
		rightmostBit, ok := rightmostPathBit(proof.Path[i-1].Path)
		if !ok {
			t.Fatalf("proof path[%d] has no meaningful bits", i-1)
		}

		if rightmostBit {
			currentHash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[i].Path, proof.Path[i].Hash, currentHash)
		} else {
			currentHash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[i].Path, currentHash, proof.Path[i].Hash)
		}
	}

	return currentHash
}

func rightmostPathBit(encodedPath []byte) (bool, bool) {
	decodedPath, trailingPadding := smt.DecodePath(encodedPath)
	meaningfulBits := len(decodedPath)*8 - trailingPadding
	if meaningfulBits <= 0 {
		return false, false
	}

	return utils.GetBit(decodedPath, 0), true
}
