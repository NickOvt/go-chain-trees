package smt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"
	"strings"

	"github.com/NickOvt/go-chain-trees/utils"
)

// Path stores a compressed SMT path segment as LSB-first bits.
// Bit 0 is the first branching decision from parent to child.
type Path struct {
	bits   big.Int
	bitLen int
}

func (path *Path) Encode() []byte {
	return encodePath(path)
}

func (path *Path) BitLen() int {
	return pathBitLen(path)
}

func (path *Path) BitAt(idx int) int {
	return pathBit(path, idx)
}

func (path *Path) TraversalBytes() ([]byte, int) {
	return pathToTraversalBytes(path)
}

func (path *Path) KeyBits() ([]byte, int) {
	bits, _ := pathToTraversalBytes(path)
	return bits, pathBitLen(path)
}

func newPathFromBigInt(bits *big.Int, bitLen int) *Path {
	if bitLen < 0 {
		bitLen = 0
	}

	path := &Path{bitLen: bitLen}
	if bits != nil {
		path.bits.Set(bits)
	}

	if bitLen == 0 {
		path.bits.SetInt64(0)
		return path
	}

	mask := new(big.Int).Lsh(big.NewInt(1), uint(bitLen))
	mask.Sub(mask, big.NewInt(1))
	path.bits.And(&path.bits, mask)

	return path
}

func clonePath(path *Path) *Path {
	if path == nil {
		return nil
	}
	return newPathFromBigInt(&path.bits, path.bitLen)
}

func pathFromTraversalBytes(data []byte, depth int) *Path {
	totalBits := depth
	if totalBits < 0 {
		totalBits = 0
	}
	if totalBits > len(data)*8 {
		totalBits = len(data) * 8
	}

	bits := new(big.Int)
	for i := 0; i < totalBits; i++ {
		if utils.GetBit(data, i) {
			bits.SetBit(bits, i, 1)
		}
	}

	return newPathFromBigInt(bits, totalBits)
}

func pathToTraversalBytes(path *Path) ([]byte, int) {
	if path == nil || path.bitLen <= 0 {
		return []byte{}, 0
	}

	result := make([]byte, (path.bitLen+7)/8)
	for i := 0; i < path.bitLen; i++ {
		if path.bits.Bit(i) == 1 {
			result[i/8] |= 1 << (7 - (i % 8))
		}
	}

	paddingBits := len(result)*8 - path.bitLen
	return result, paddingBits
}

func pathFromKeyBytes(key []byte, bitLen int) *Path {
	if bitLen <= 0 {
		bitLen = len(key) * 8
	}
	return newPathFromBigInt(new(big.Int).SetBytes(key), bitLen)
}

func pathBit(path *Path, idx int) int {
	if path == nil || idx < 0 || idx >= path.bitLen {
		return 0
	}
	return int(path.bits.Bit(idx))
}

func pathBitLen(path *Path) int {
	if path == nil {
		return 0
	}
	return path.bitLen
}

func pathEqual(a, b *Path) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.bitLen == b.bitLen && a.bits.Cmp(&b.bits) == 0
}

func pathCutPrefix(path *Path, prefixBits int) *Path {
	if path == nil {
		return nil
	}
	if prefixBits <= 0 {
		return clonePath(path)
	}
	if prefixBits >= path.bitLen {
		return newPathFromBigInt(nil, 0)
	}

	shifted := new(big.Int).Rsh(&path.bits, uint(prefixBits))
	return newPathFromBigInt(shifted, path.bitLen-prefixBits)
}

func pathCommonPrefix(a, b *Path) (*Path, int) {
	if a == nil || b == nil {
		return newPathFromBigInt(nil, 0), 0
	}

	maxBits := a.bitLen
	if b.bitLen < maxBits {
		maxBits = b.bitLen
	}

	prefixLen := 0
	for prefixLen < maxBits {
		if a.bits.Bit(prefixLen) != b.bits.Bit(prefixLen) {
			break
		}
		prefixLen++
	}

	if prefixLen == 0 {
		return newPathFromBigInt(nil, 0), 0
	}

	mask := new(big.Int).Lsh(big.NewInt(1), uint(prefixLen))
	mask.Sub(mask, big.NewInt(1))
	prefixBits := new(big.Int).And(&a.bits, mask)

	return newPathFromBigInt(prefixBits, prefixLen), prefixLen
}

func encodePath(path *Path) []byte {
	if path == nil {
		return nil
	}

	byteLen := (path.bitLen + 7) / 8
	encoded := make([]byte, binary.MaxVarintLen64+byteLen)
	n := binary.PutUvarint(encoded, uint64(path.bitLen))
	encoded = encoded[:n+byteLen]

	for i := 0; i < byteLen; i++ {
		var b byte
		for bit := 0; bit < 8; bit++ {
			idx := i*8 + bit
			if idx >= path.bitLen {
				break
			}
			if path.bits.Bit(idx) == 1 {
				b |= 1 << bit
			}
		}
		encoded[n+i] = b
	}

	return encoded
}

func decodeEncodedPath(encoded []byte) (*Path, bool) {
	if len(encoded) == 0 {
		return newPathFromBigInt(nil, 0), true
	}

	bitLen, n := binary.Uvarint(encoded)
	if n <= 0 {
		return nil, false
	}

	byteLen := int((bitLen + 7) / 8)
	if len(encoded) < n+byteLen {
		return nil, false
	}

	bits := new(big.Int)
	for i := 0; i < byteLen; i++ {
		b := encoded[n+i]
		for bit := 0; bit < 8; bit++ {
			idx := i*8 + bit
			if idx >= int(bitLen) {
				break
			}
			if (b & (1 << bit)) != 0 {
				bits.SetBit(bits, idx, 1)
			}
		}
	}

	return newPathFromBigInt(bits, int(bitLen)), true
}

func pathToRawBits(path *Path) string {
	if path == nil || path.bitLen <= 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(path.bitLen)
	for i := path.bitLen - 1; i >= 0; i-- {
		if path.bits.Bit(i) == 1 {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}

	return sb.String()
}

func nodePathBytes(path *Path) []byte {
	return encodePath(path)
}

func EncodePath(data []byte, depth int) ([]byte, int) {
	path := pathFromTraversalBytes(data, depth)
	return encodePath(path), pathBitLen(path)
}

// left aligned, padded with 0 in result
func DecodePath(encoded []byte) ([]byte, int) {
	path, ok := decodeEncodedPath(encoded)
	if !ok {
		return []byte{}, 0
	}
	return pathToTraversalBytes(path)
}

// CalculateKeyFromPath decodes an encoded path and returns traversal-order bits.
// Bit index 0 in the returned byte array is the first edge decision from parent.
func CalculateKeyFromPath(encodedPath []byte) ([]byte, int) {
	if encodedPath == nil {
		return []byte{}, 0
	}

	path, ok := decodeEncodedPath(encodedPath)
	if !ok {
		return []byte{}, 0
	}

	decoded, _ := pathToTraversalBytes(path)
	return decoded, pathBitLen(path)
}

// EncodeKeyBitsAsPath encodes traversal-order bits into canonical SMT path bytes.
func EncodeKeyBitsAsPath(keyBits []byte, meaningfulBits int) []byte {
	encoded, _ := EncodePath(keyBits, meaningfulBits)
	return encoded
}

type SMT struct {
	HashAlgo       utils.HashAlgo
	Root           *Node
	emptyHashCache map[int]utils.Hash
	emptyHash      utils.Hash
	AppendOnly     bool
}

type Node struct {
	Key       utils.Hash     // Key of node (will be hashed by the tree hashAlgo), nil for branch node
	Data      utils.CBORData // in case of a non-leaf node it will be nil
	Hash      utils.Hash     // present on every node. hash(path, data), for branch nodes hash(path, leftHash, rightHash)
	Path      *Path          // nil only for root node
	LeftNode  *Node          // For branch nodes
	RightNode *Node          // For branch nodes
	IsLeaf    bool
}

type ProofNode struct {
	Path []byte
	Hash utils.Hash
	Data []byte
}

type InclusionExclusionProof struct {
	Root utils.Hash
	Path []*ProofNode
}

type PublicInclusionExclusionProof struct {
	Root utils.Hash  `cbor:"1,keyasint" json:"root"`
	Path [][2][]byte `cbor:"2,keyasint" json:"path"`
}

func (proof *InclusionExclusionProof) ToPublicProof() *PublicInclusionExclusionProof {
	if proof == nil {
		return nil
	}

	path := make([][2][]byte, len(proof.Path))
	for i, proofNode := range proof.Path {
		if proofNode == nil {
			continue
		}

		path[i][0] = slices.Clone(proofNode.Path)
		if len(proofNode.Data) > 0 {
			path[i][1] = slices.Clone(proofNode.Data)
		} else {
			path[i][1] = slices.Clone(proofNode.Hash)
		}
	}

	return &PublicInclusionExclusionProof{
		Root: slices.Clone(proof.Root),
		Path: path,
	}
}

func (proof *PublicInclusionExclusionProof) ToInternalProof() *InclusionExclusionProof {
	if proof == nil {
		return nil
	}

	path := make([]*ProofNode, len(proof.Path))
	for i, tuple := range proof.Path {
		proofNode := &ProofNode{
			Path: slices.Clone(tuple[0]),
			Hash: nil,
			Data: nil,
		}

		if i == 0 {
			proofNode.Data = slices.Clone(tuple[1])
		} else {
			proofNode.Hash = slices.Clone(tuple[1])
		}

		path[i] = proofNode
	}

	return &InclusionExclusionProof{
		Root: slices.Clone(proof.Root),
		Path: path,
	}
}

func (t *SMT) VerifyPublicProof(proof *PublicInclusionExclusionProof) (bool, error) {
	return t.VerifyProof(proof.ToInternalProof())
}

func (t *SMT) VerifyProof(proof *InclusionExclusionProof) (bool, error) {
	if proof == nil {
		return false, nil
	}

	if len(proof.Path) == 0 {
		return false, fmt.Errorf("proof path is empty")
	}
	if len(proof.Path) < 2 {
		return false, fmt.Errorf("proof path must contain at least leaf and parent witnesses")
	}

	if !bytes.Equal(proof.Root, t.GetRoot().GetHash()) {
		return false, fmt.Errorf("given proof document contains invalid root hash")
	}

	recomputedRoot, err := recomputeRootFromProofBySpec(t.HashAlgo, proof)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(recomputedRoot, proof.Root) {
		return false, fmt.Errorf("calculated root hash does not match provided root hash")
	}

	if err := t.verifyLeafPathMatchesLeafKey(proof); err != nil {
		return false, err
	}

	return true, nil
}

func buildFullLeafPathFromProof(proof *InclusionExclusionProof) (*Path, error) {
	if proof == nil || len(proof.Path) == 0 {
		return nil, fmt.Errorf("cannot build leaf path from empty proof")
	}

	combined := new(big.Int)
	totalBits := 0

	for i := len(proof.Path) - 1; i >= 0; i-- {
		if proof.Path[i] == nil {
			return nil, fmt.Errorf("proof contains nil node at index %d", i)
		}

		segment, ok := decodeEncodedPath(proof.Path[i].Path)
		if !ok {
			return nil, fmt.Errorf("proof path[%d] is not a valid encoded path", i)
		}

		if pathBitLen(segment) == 0 {
			continue
		}

		shifted := new(big.Int).Lsh(new(big.Int).Set(&segment.bits), uint(totalBits))
		combined.Or(combined, shifted)
		totalBits += pathBitLen(segment)
	}

	return newPathFromBigInt(combined, totalBits), nil
}

func (t *SMT) findLeafByFullPath(fullPath *Path) (*Node, error) {
	if fullPath == nil {
		return nil, fmt.Errorf("full path is nil")
	}

	current := t.GetRoot()
	remaining := clonePath(fullPath)
	depth := 0

	for current != nil && !current.IsLeaf {
		if pathBitLen(remaining) <= 0 {
			return nil, fmt.Errorf("path exhausted before reaching leaf")
		}

		goLeft := pathBit(remaining, 0) == 0

		var child *Node
		if goLeft {
			child = current.LeftNode
		} else {
			child = current.RightNode
		}
		if child == nil {
			return nil, fmt.Errorf("missing child for path at depth %d", depth)
		}

		_, commonPrefixLen := pathCommonPrefix(child.Path, remaining)
		if commonPrefixLen != pathBitLen(child.Path) {
			return nil, fmt.Errorf("path mismatch at depth %d", depth)
		}

		remaining = pathCutPrefix(remaining, commonPrefixLen)
		depth += commonPrefixLen
		current = child
	}

	if current == nil || !current.IsLeaf {
		return nil, fmt.Errorf("leaf not found for calculated path")
	}

	if pathBitLen(remaining) != 0 {
		return nil, fmt.Errorf("calculated path has extra bits past located leaf")
	}

	return current, nil
}

func (t *SMT) verifyLeafPathMatchesLeafKey(proof *InclusionExclusionProof) error {
	fullPath, err := buildFullLeafPathFromProof(proof)
	if err != nil {
		return err
	}

	keyBitLen := utils.GetHashAlgoOutputBitCount(t.HashAlgo)
	if pathBitLen(fullPath) != keyBitLen {
		return fmt.Errorf("calculated leaf path bit length %d does not match key bit length %d", pathBitLen(fullPath), keyBitLen)
	}

	leaf, err := t.findLeafByFullPath(fullPath)
	if err != nil {
		return err
	}
	if len(leaf.Key) == 0 {
		return fmt.Errorf("located leaf has empty key")
	}

	leafKeyPath := pathFromKeyBytes(leaf.Key, len(leaf.Key)*8)
	if !pathEqual(leafKeyPath, fullPath) {
		return fmt.Errorf("calculated leaf path does not match leaf key")
	}

	return nil
}

func recomputeRootFromProofBySpec(hashAlgo utils.HashAlgo, proof *InclusionExclusionProof) (utils.Hash, error) {
	if proof == nil || len(proof.Path) == 0 {
		return nil, fmt.Errorf("cannot recompute root from empty proof")
	}
	if len(proof.Path[0].Data) == 0 {
		return nil, fmt.Errorf("first proof node must contain leaf data for inclusion verification")
	}

	currentHash := utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[0].Path, proof.Path[0].Data)

	for i := 1; i < len(proof.Path); i++ {
		if proof.Path[i] == nil {
			return nil, fmt.Errorf("proof contains nil node at index %d", i)
		}

		firstBit, ok := firstEncodedPathBit(proof.Path[i-1].Path)
		if !ok {
			return nil, fmt.Errorf("proof path[%d] has no meaningful bits", i-1)
		}

		if firstBit {
			currentHash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[i].Path, proof.Path[i].Hash, currentHash)
		} else {
			currentHash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, proof.Path[i].Path, currentHash, proof.Path[i].Hash)
		}
	}

	return currentHash, nil
}

func firstEncodedPathBit(encodedPath []byte) (bool, bool) {
	path, ok := decodeEncodedPath(encodedPath)
	if !ok || pathBitLen(path) <= 0 {
		return false, false
	}

	return pathBit(path, 0) == 1, true
}

func (node *Node) GetLeftNode() *Node {
	if node == nil {
		return nil
	}
	return node.LeftNode
}

func (node *Node) GetRightNode() *Node {
	if node == nil {
		return nil
	}
	return node.RightNode
}

func (node *Node) GetHash() utils.Hash {
	if node == nil {
		return nil
	}
	return node.Hash
}

func (node *Node) CalculateLeafHash(hashAlgo utils.HashAlgo) {
	if node == nil {
		return
	}

	node.Hash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, nodePathBytes(node.Path), node.Data)
}

func (node *Node) CalculateBranchHash(hashAlgo utils.HashAlgo) {
	if node == nil {
		return
	}

	node.Hash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, nodePathBytes(node.Path), node.GetLeftNode().GetHash(), node.GetRightNode().GetHash())
}

func NewSMT(hashAlgo utils.HashAlgo, appendOnly bool) *SMT {
	smt := &SMT{HashAlgo: hashAlgo, AppendOnly: appendOnly}
	smt.Init()
	return smt
}

func (t *SMT) Init() {
	if t.emptyHash == nil {
		t.emptyHash = utils.GenerateNullHash(t.HashAlgo)
	}

	// set initial values
	rootNode := &Node{
		Data:      nil,
		LeftNode:  nil,
		RightNode: nil,
		Path:      nil,
		Hash:      utils.ConcatHashesAndGenerateHash(t.HashAlgo, nil, nil, nil),
		IsLeaf:    false,
	}
	t.Root = rootNode

	t.emptyHashCache = make(map[int]utils.Hash)
}

func (t *SMT) GetRoot() *Node {
	if t.Root == nil {
		rootNode := &Node{
			Data:      nil,
			LeftNode:  nil,
			RightNode: nil,
			Path:      nil,
			Hash:      utils.ConcatHashesAndGenerateHash(t.HashAlgo, nil, nil, nil),
			IsLeaf:    false,
		}
		t.Root = rootNode
	}

	return t.Root
}

func (t *SMT) Insert(key []byte, data []byte) (bool, error) {
	_, err := t.insert(key, data, t.GetRoot(), nil, t.AppendOnly)

	if err != nil {
		return false, err
	}

	return true, nil
}

// PrintTree writes a visual representation of the SMT to stdout.
// Branch nodes display their decoded raw path bitstrings, while leaf nodes
// display their hashes (and paths too).
func (t *SMT) PrintTree() {
	t.WriteTree(os.Stdout)
}

// WriteTree writes a visual representation of the SMT to the provided writer.
// Branch nodes display their decoded raw path bitstrings, while leaf nodes
// display their hashes (and paths too).
func (t *SMT) WriteTree(w io.Writer) {
	if w == nil {
		return
	}

	if t == nil || t.GetRoot() == nil {
		fmt.Fprintln(w, "SMT is empty")
		return
	}

	root := t.GetRoot()
	fmt.Fprintln(w, "SMT Tree Structure:")
	fmt.Fprintf(w, "root: branch path=%s\n", formatBranchPath(root.Path))

	children := make([]struct {
		label string
		node  *Node
	}, 0, 2)
	if root.LeftNode != nil {
		children = append(children, struct {
			label string
			node  *Node
		}{label: "L", node: root.LeftNode})
	}
	if root.RightNode != nil {
		children = append(children, struct {
			label string
			node  *Node
		}{label: "R", node: root.RightNode})
	}

	for i, child := range children {
		isLast := i == len(children)-1
		writeTreeNode(w, child.node, "", isLast, child.label)
	}
}

func writeTreeNode(w io.Writer, node *Node, prefix string, isLast bool, edgeLabel string) {
	if node == nil {
		return
	}

	connector := "|--"
	nextPrefix := prefix + "|   "
	if isLast {
		connector = "`--"
		nextPrefix = prefix + "    "
	}

	if node.IsLeaf {
		fmt.Fprintf(w, "%s%s %s: leaf path=%s hash=%x\n", prefix, connector, edgeLabel, formatLeafPath(node.Path), node.Hash)
		return
	}

	fmt.Fprintf(w, "%s%s %s: branch path=%s\n", prefix, connector, edgeLabel, formatBranchPath(node.Path))

	children := make([]struct {
		label string
		node  *Node
	}, 0, 2)
	if node.LeftNode != nil {
		children = append(children, struct {
			label string
			node  *Node
		}{label: "L", node: node.LeftNode})
	}
	if node.RightNode != nil {
		children = append(children, struct {
			label string
			node  *Node
		}{label: "R", node: node.RightNode})
	}

	for i, child := range children {
		childIsLast := i == len(children)-1
		writeTreeNode(w, child.node, nextPrefix, childIsLast, child.label)
	}
}

func formatBranchPath(path *Path) string {
	if path == nil {
		return "<root>"
	}

	bits := decodePathToRawBits(path)
	if bits == "" {
		return "<empty>"
	}

	return bits
}

func formatLeafPath(path *Path) string {
	bits := decodePathToRawBits(path)
	if bits == "" {
		return "<empty>"
	}

	return bits
}

func decodePathToRawBits(path *Path) string {
	return pathToRawBits(path)
}

// ChooseNewKey Returns both the key and the number of meaningful bits for correct comparison.
func ChooseNewKey(originalKeyHash utils.Hash, newEncodedPath []byte) ([]byte, int) {
	if newEncodedPath == nil || len(newEncodedPath) == 0 {
		return originalKeyHash, len(originalKeyHash) * 8
	}

	return CalculateKeyFromPath(newEncodedPath)
}

func chooseNewPath(originalKeyHash utils.Hash, newPath *Path) *Path {
	if newPath == nil {
		return pathFromKeyBytes(originalKeyHash, len(originalKeyHash)*8)
	}
	return clonePath(newPath)
}

func newLeafNode(hashAlgo utils.HashAlgo, keyHash utils.Hash, data utils.CBORData, path *Path) *Node {
	leafPath := clonePath(path)
	encodedPath := nodePathBytes(leafPath)

	return &Node{
		Data:      data,
		LeftNode:  nil,
		RightNode: nil,
		Path:      leafPath,
		Hash:      utils.ConcatDataAndGenerateCombinedHash(hashAlgo, encodedPath, data),
		IsLeaf:    true,
		Key:       slices.Clone(keyHash),
	}
}

func recalculateNodeHash(node *Node, hashAlgo utils.HashAlgo) {
	if node == nil {
		return
	}
	if node.IsLeaf {
		node.CalculateLeafHash(hashAlgo)
		return
	}
	node.CalculateBranchHash(hashAlgo)
}

func (t *SMT) insert(key []byte, data []byte, currentRoot *Node, newPath *Path, appendOnly bool) (*Node, error) {
	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
	nodeDataCbor, err := utils.EncodeCBOR(data)
	if err != nil {
		return nil, err
	}

	chosenPath := chooseNewPath(nodeKeyHash, newPath)
	if pathBitLen(chosenPath) <= 0 {
		return nil, fmt.Errorf("cannot insert with empty path")
	}

	goLeft := pathBit(chosenPath, 0) == 0

	if goLeft {
		if currentRoot.LeftNode == nil {
			currentRoot.LeftNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, chosenPath)
			currentRoot.CalculateBranchHash(t.HashAlgo)
			return currentRoot, nil
		}

		if currentRoot.LeftNode.IsLeaf {
			if bytes.Equal(currentRoot.LeftNode.Key, nodeKeyHash) {
				if appendOnly {
					return nil, fmt.Errorf("duplicate key insert in append-only mode")
				}
				currentRoot.LeftNode.Data = nodeDataCbor
				currentRoot.LeftNode.CalculateLeafHash(t.HashAlgo)
				currentRoot.CalculateBranchHash(t.HashAlgo)
				return currentRoot, nil
			}

			existingPath := clonePath(currentRoot.LeftNode.Path)
			commonPrefix, commonPrefixLen := pathCommonPrefix(existingPath, chosenPath)

			branchNode := &Node{
				Data:      nil,
				LeftNode:  nil,
				RightNode: nil,
				Path:      commonPrefix,
				IsLeaf:    false,
				Key:       nil,
			}

			newNodePath := pathCutPrefix(chosenPath, commonPrefixLen)
			oldNodePath := pathCutPrefix(existingPath, commonPrefixLen)

			currentRoot.LeftNode.Path = oldNodePath
			currentRoot.LeftNode.CalculateLeafHash(t.HashAlgo)

			if pathBit(existingPath, commonPrefixLen) == 1 {
				branchNode.RightNode = currentRoot.LeftNode
				branchNode.LeftNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
			} else {
				branchNode.LeftNode = currentRoot.LeftNode
				branchNode.RightNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
			}

			branchNode.CalculateBranchHash(t.HashAlgo)
			currentRoot.LeftNode = branchNode
			currentRoot.CalculateBranchHash(t.HashAlgo)
			return currentRoot, nil
		}

		existingPath := clonePath(currentRoot.LeftNode.Path)
		commonPrefix, commonPrefixLen := pathCommonPrefix(existingPath, chosenPath)

		if pathEqual(commonPrefix, currentRoot.LeftNode.Path) {
			newPathForRecursion := pathCutPrefix(chosenPath, commonPrefixLen)
			if _, err := t.insert(key, data, currentRoot.LeftNode, newPathForRecursion, appendOnly); err != nil {
				return nil, err
			}
			currentRoot.CalculateBranchHash(t.HashAlgo)
			return currentRoot, nil
		}

		newBranchNode := &Node{
			Data:      nil,
			LeftNode:  nil,
			RightNode: nil,
			Path:      commonPrefix,
			IsLeaf:    false,
			Key:       nil,
		}

		newNodePath := pathCutPrefix(chosenPath, commonPrefixLen)
		oldBranchPath := pathCutPrefix(existingPath, commonPrefixLen)

		currentRoot.LeftNode.Path = oldBranchPath
		recalculateNodeHash(currentRoot.LeftNode, t.HashAlgo)

		if pathBit(existingPath, commonPrefixLen) == 1 {
			newBranchNode.RightNode = currentRoot.LeftNode
			newBranchNode.LeftNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
		} else {
			newBranchNode.LeftNode = currentRoot.LeftNode
			newBranchNode.RightNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
		}

		newBranchNode.CalculateBranchHash(t.HashAlgo)
		currentRoot.LeftNode = newBranchNode
		currentRoot.CalculateBranchHash(t.HashAlgo)
		return currentRoot, nil
	}

	if currentRoot.RightNode == nil {
		currentRoot.RightNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, chosenPath)
		currentRoot.CalculateBranchHash(t.HashAlgo)
		return currentRoot, nil
	}

	if currentRoot.RightNode.IsLeaf {
		if bytes.Equal(currentRoot.RightNode.Key, nodeKeyHash) {
			if appendOnly {
				return nil, fmt.Errorf("duplicate key insert in append-only mode")
			}
			currentRoot.RightNode.Data = nodeDataCbor
			currentRoot.RightNode.CalculateLeafHash(t.HashAlgo)
			currentRoot.CalculateBranchHash(t.HashAlgo)
			return currentRoot, nil
		}

		existingPath := clonePath(currentRoot.RightNode.Path)
		commonPrefix, commonPrefixLen := pathCommonPrefix(existingPath, chosenPath)

		branchNode := &Node{
			Data:      nil,
			LeftNode:  nil,
			RightNode: nil,
			Path:      commonPrefix,
			IsLeaf:    false,
			Key:       nil,
		}

		newNodePath := pathCutPrefix(chosenPath, commonPrefixLen)
		oldNodePath := pathCutPrefix(existingPath, commonPrefixLen)

		currentRoot.RightNode.Path = oldNodePath
		currentRoot.RightNode.CalculateLeafHash(t.HashAlgo)

		if pathBit(existingPath, commonPrefixLen) == 1 {
			branchNode.RightNode = currentRoot.RightNode
			branchNode.LeftNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
		} else {
			branchNode.LeftNode = currentRoot.RightNode
			branchNode.RightNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
		}

		branchNode.CalculateBranchHash(t.HashAlgo)
		currentRoot.RightNode = branchNode
		currentRoot.CalculateBranchHash(t.HashAlgo)
		return currentRoot, nil
	}

	existingPath := clonePath(currentRoot.RightNode.Path)
	commonPrefix, commonPrefixLen := pathCommonPrefix(existingPath, chosenPath)

	if pathEqual(commonPrefix, currentRoot.RightNode.Path) {
		newPathForRecursion := pathCutPrefix(chosenPath, commonPrefixLen)
		if _, err := t.insert(key, data, currentRoot.RightNode, newPathForRecursion, appendOnly); err != nil {
			return nil, err
		}
		currentRoot.CalculateBranchHash(t.HashAlgo)
		return currentRoot, nil
	}

	newBranchNode := &Node{
		Data:      nil,
		LeftNode:  nil,
		RightNode: nil,
		Path:      commonPrefix,
		IsLeaf:    false,
		Key:       nil,
	}

	newNodePath := pathCutPrefix(chosenPath, commonPrefixLen)
	oldBranchPath := pathCutPrefix(existingPath, commonPrefixLen)

	currentRoot.RightNode.Path = oldBranchPath
	recalculateNodeHash(currentRoot.RightNode, t.HashAlgo)

	if pathBit(existingPath, commonPrefixLen) == 1 {
		newBranchNode.RightNode = currentRoot.RightNode
		newBranchNode.LeftNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
	} else {
		newBranchNode.LeftNode = currentRoot.RightNode
		newBranchNode.RightNode = newLeafNode(t.HashAlgo, nodeKeyHash, nodeDataCbor, newNodePath)
	}

	newBranchNode.CalculateBranchHash(t.HashAlgo)
	currentRoot.RightNode = newBranchNode
	currentRoot.CalculateBranchHash(t.HashAlgo)

	return currentRoot, nil
}

func (t *SMT) GenerateInclusionExclusionProof(key []byte) (*InclusionExclusionProof, error) {
	rootNode := t.GetRoot()
	proof := &InclusionExclusionProof{
		Path: []*ProofNode{},
		Root: rootNode.GetHash(),
	}

	if rootNode.LeftNode == nil && rootNode.RightNode == nil {
		return nil, fmt.Errorf("Cannot generate inclusion proof for an empty tree")
	}

	keyBitSize := utils.GetHashAlgoOutputBitCount(t.HashAlgo)
	fullKeyPath := pathFromKeyBytes(key, keyBitSize)

	i := 0
	currNode := rootNode
	pathMismatch := false

	for i < keyBitSize && currNode != nil && !currNode.IsLeaf {
		keyAtDepth := pathCutPrefix(fullKeyPath, i)
		if pathBitLen(keyAtDepth) <= 0 {
			break
		}

		goLeft := pathBit(keyAtDepth, 0) == 0

		proofNode := &ProofNode{
			Path: nodePathBytes(currNode.Path),
			Data: nil,
		}

		var nextNode *Node
		if goLeft {
			if currNode.RightNode != nil {
				proofNode.Hash = currNode.RightNode.GetHash()
			}
			nextNode = currNode.LeftNode
		} else {
			if currNode.LeftNode != nil {
				proofNode.Hash = currNode.LeftNode.GetHash()
			}
			nextNode = currNode.RightNode
		}

		proof.Path = append(proof.Path, proofNode)

		if nextNode == nil {
			currNode = nil
			break
		}

		_, commonPrefixLen := pathCommonPrefix(nextNode.Path, keyAtDepth)
		i += commonPrefixLen
		currNode = nextNode

		if commonPrefixLen != pathBitLen(currNode.Path) {
			pathMismatch = true
			break
		}
	}

	if currNode != nil {
		if currNode.IsLeaf {
			proof.Path = append(proof.Path, &ProofNode{
				Path: nodePathBytes(currNode.Path),
				Hash: nil,
				Data: currNode.Data,
			})
		} else if pathMismatch || i >= keyBitSize {
			proof.Path = append(proof.Path, &ProofNode{
				Path: nodePathBytes(currNode.Path),
				Hash: currNode.Hash,
				Data: nil,
			})
		}
	}

	slices.Reverse(proof.Path)
	return proof, nil
}
