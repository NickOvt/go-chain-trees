package smt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/bits"
	"os"
	"slices"
	"strings"

	"github.com/NickOvt/go-chain-trees/utils"
)

// Path stores a compressed SMT path segment as LSB-first bits.
// Bit 0 is the first branching decision from parent to child.
type Path struct {
	bits   []byte // LSB-packed: bit i is bits[i/8]&(1<<(i%8))
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

func pathByteLen(bitLen int) int {
	if bitLen <= 0 {
		return 0
	}

	return (bitLen + 7) / 8
}

func maskUnusedPathBits(raw []byte, bitLen int) {
	if bitLen <= 0 || len(raw) == 0 {
		return
	}

	lastByteUsedBits := bitLen & 7
	if lastByteUsedBits == 0 {
		return
	}

	raw[len(raw)-1] &= byte((1 << lastByteUsedBits) - 1)
}

func newPathFromBits(raw []byte, bitLen int) *Path {
	if bitLen < 0 {
		bitLen = 0
	}

	byteLen := pathByteLen(bitLen)
	path := &Path{
		bits:   make([]byte, byteLen),
		bitLen: bitLen,
	}
	copy(path.bits, raw)

	maskUnusedPathBits(path.bits, bitLen)

	return path
}

func clonePath(path *Path) *Path {
	if path == nil {
		return nil
	}

	return newPathFromBits(path.bits, path.bitLen)
}

func pathFromTraversalBytes(data []byte, depth int) *Path {
	totalBits := depth
	if totalBits < 0 {
		totalBits = 0
	}
	if totalBits > len(data)*8 {
		totalBits = len(data) * 8
	}

	byteLen := pathByteLen(totalBits)
	pathBits := make([]byte, byteLen)
	for i := 0; i < byteLen; i++ {
		pathBits[i] = bits.Reverse8(data[i])
	}
	maskUnusedPathBits(pathBits, totalBits)

	return &Path{
		bits:   pathBits,
		bitLen: totalBits,
	}
}

func pathFromKeyBytes(key []byte, bitLen int) *Path {
	if bitLen <= 0 {
		bitLen = len(key) * 8
	}

	byteLen := pathByteLen(bitLen)
	pathBits := make([]byte, byteLen)

	maxBytesToCopy := len(key)
	if maxBytesToCopy > byteLen {
		maxBytesToCopy = byteLen
	}
	for i := 0; i < maxBytesToCopy; i++ {
		pathBits[i] = key[len(key)-1-i]
	}
	maskUnusedPathBits(pathBits, bitLen)

	return &Path{
		bits:   pathBits,
		bitLen: bitLen,
	}
}

// keyPathBit returns path bit `idx` derived from key bytes without allocating a Path.
// Path bit ordering matches pathFromKeyBytes + pathBit (LSB-first within each reversed byte).
func keyPathBit(key []byte, idx int) int {
	if idx < 0 || idx >= len(key)*8 {
		return 0
	}

	byteIdx := len(key) - 1 - (idx >> 3)
	if byteIdx < 0 || byteIdx >= len(key) {
		return 0
	}

	return int((key[byteIdx] >> (idx & 7)) & 1)
}

func pathBit(path *Path, idx int) int {
	if path == nil || idx < 0 || idx >= path.bitLen {
		return 0
	}

	byteIdx := idx >> 3
	if byteIdx >= len(path.bits) {
		return 0
	}

	return int((path.bits[byteIdx] >> (idx & 7)) & 1)
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

	return a.bitLen == b.bitLen && bytes.Equal(a.bits, b.bits)
}

func copyPathBitsShifted(dst, src []byte, shiftBits int) {
	if len(dst) == 0 {
		return
	}

	byteShift := shiftBits >> 3
	bitShift := shiftBits & 7

	if bitShift == 0 {
		copy(dst, src[byteShift:byteShift+len(dst)])
		return
	}

	for i := range dst {
		srcIdx := byteShift + i
		b := src[srcIdx] >> bitShift
		if srcIdx+1 < len(src) {
			b |= src[srcIdx+1] << (8 - bitShift)
		}
		dst[i] = b
	}
}

func pathCutPrefix(path *Path, prefixBits int) *Path {
	if path == nil {
		return nil
	}
	if prefixBits <= 0 {
		return clonePath(path)
	}
	if prefixBits >= path.bitLen {
		return &Path{}
	}

	resultBitLen := path.bitLen - prefixBits
	result := &Path{
		bits:   make([]byte, pathByteLen(resultBitLen)),
		bitLen: resultBitLen,
	}
	copyPathBitsShifted(result.bits, path.bits, prefixBits)
	maskUnusedPathBits(result.bits, result.bitLen)

	return result
}

func pathSlice(path *Path, startBit int, bitLen int) *Path {
	if path == nil || bitLen <= 0 {
		return &Path{}
	}
	if startBit < 0 {
		startBit = 0
	}
	if startBit >= path.bitLen {
		return &Path{}
	}

	remaining := path.bitLen - startBit
	if bitLen > remaining {
		bitLen = remaining
	}
	if bitLen <= 0 {
		return &Path{}
	}

	result := &Path{
		bits:   make([]byte, pathByteLen(bitLen)),
		bitLen: bitLen,
	}
	copyPathBitsShifted(result.bits, path.bits, startBit)
	maskUnusedPathBits(result.bits, result.bitLen)

	return result
}

func pathCommonPrefixLenAt(a *Path, aOffset int, b *Path, bOffset int) int {
	if a == nil || b == nil {
		return 0
	}
	if aOffset < 0 {
		aOffset = 0
	}
	if bOffset < 0 {
		bOffset = 0
	}
	if aOffset >= a.bitLen || bOffset >= b.bitLen {
		return 0
	}

	maxBits := a.bitLen - aOffset
	if bRemaining := b.bitLen - bOffset; bRemaining < maxBits {
		maxBits = bRemaining
	}

	prefixLen := 0
	for prefixLen < maxBits && (((aOffset+prefixLen)&7) != 0 || ((bOffset+prefixLen)&7) != 0) {
		if pathBit(a, aOffset+prefixLen) != pathBit(b, bOffset+prefixLen) {
			return prefixLen
		}
		prefixLen++
	}

	for prefixLen+8 <= maxBits {
		aByteIdx := (aOffset + prefixLen) >> 3
		bByteIdx := (bOffset + prefixLen) >> 3
		x := a.bits[aByteIdx] ^ b.bits[bByteIdx]
		if x == 0 {
			prefixLen += 8
			continue
		}

		prefixLen += bits.TrailingZeros8(x)
		return prefixLen
	}

	for prefixLen < maxBits {
		if pathBit(a, aOffset+prefixLen) != pathBit(b, bOffset+prefixLen) {
			break
		}
		prefixLen++
	}

	return prefixLen
}

func pathCommonPrefixLenAtKey(path *Path, pathOffset int, key []byte, keyOffset int, keyBitLen int) int {
	if path == nil || len(key) == 0 {
		return 0
	}
	if pathOffset < 0 {
		pathOffset = 0
	}
	if keyOffset < 0 {
		keyOffset = 0
	}
	if keyBitLen <= 0 {
		keyBitLen = len(key) * 8
	}
	if pathOffset >= path.bitLen || keyOffset >= keyBitLen {
		return 0
	}

	maxBits := path.bitLen - pathOffset
	if keyRemaining := keyBitLen - keyOffset; keyRemaining < maxBits {
		maxBits = keyRemaining
	}

	prefixLen := 0
	for prefixLen < maxBits {
		if pathBit(path, pathOffset+prefixLen) != keyPathBit(key, keyOffset+prefixLen) {
			break
		}
		prefixLen++
	}

	return prefixLen
}

func pathCommonPrefix(a, b *Path) (*Path, int) {
	prefixLen := pathCommonPrefixLenAt(a, 0, b, 0)
	if prefixLen == 0 {
		return &Path{}, 0
	}

	return pathSlice(a, 0, prefixLen), prefixLen
}

func encodePath(path *Path) []byte {
	if path == nil {
		return nil
	}

	byteLen := pathByteLen(path.bitLen)
	encoded := make([]byte, binary.MaxVarintLen64+byteLen)
	n := binary.PutUvarint(encoded, uint64(path.bitLen))
	encoded = encoded[:n+byteLen]
	copy(encoded[n:], path.bits)

	return encoded
}

func decodeEncodedPath(encoded []byte) (*Path, bool) {
	if len(encoded) == 0 {
		return &Path{}, true
	}

	bitLen, n := binary.Uvarint(encoded)
	if n <= 0 {
		return nil, false
	}

	byteLen := int((bitLen + 7) / 8)
	if len(encoded) < n+byteLen {
		return nil, false
	}

	path := &Path{
		bits:   make([]byte, byteLen),
		bitLen: int(bitLen),
	}
	copy(path.bits, encoded[n:n+byteLen])
	maskUnusedPathBits(path.bits, path.bitLen)

	return path, true
}

func pathToTraversalBytes(path *Path) ([]byte, int) {
	if path == nil || path.bitLen <= 0 {
		return []byte{}, 0
	}

	byteLen := pathByteLen(path.bitLen)
	result := make([]byte, byteLen)
	for i := 0; i < byteLen; i++ {
		result[i] = bits.Reverse8(path.bits[i])
	}

	paddingBits := len(result)*8 - path.bitLen
	return result, paddingBits
}

func pathToRawBits(path *Path) string {
	if path == nil || path.bitLen <= 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(path.bitLen)
	for i := path.bitLen - 1; i >= 0; i-- {
		if pathBit(path, i) == 1 {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}

	return sb.String()
}

func (t *SMT) calculateLeafNodeHash(encodedPath []byte, data []byte, dst utils.Hash) utils.Hash {
	if t.dataHasher == nil {
		t.dataHasher = utils.NewDataHasher(t.HashAlgo)
	}
	return t.dataHasher.Sum2To(dst, encodedPath, data)
}

func (t *SMT) calculateBranchNodeHash(encodedPath []byte, leftHash, rightHash utils.Hash, dst utils.Hash) utils.Hash {
	if t.dataHasher == nil {
		t.dataHasher = utils.NewDataHasher(t.HashAlgo)
	}
	return t.dataHasher.Sum3To(dst, encodedPath, leftHash, rightHash)
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
	HashAlgo   utils.HashAlgo
	Root       *Node
	AppendOnly bool
	dataHasher *utils.DataHasher
}

type Node struct {
	Key         utils.Hash // Key of node (will be hashed by the tree hashAlgo), nil for branch node
	Data        []byte     // Raw leaf payload bytes, nil for branch nodes
	Hash        utils.Hash // present on every node. hash(CBOR[path, data]), for branch nodes hash(CBOR[path, leftHash, rightHash])
	Path        *Path      // nil only for root node
	encodedPath []byte
	LeftNode    *Node
	RightNode   *Node
	IsLeaf      bool
}

type ProofNode struct {
	Path []byte
	Hash utils.Hash
	Data []byte
}

type InclusionExclusionProof struct {
	Root utils.Hash
	Key  utils.Hash
	Path []*ProofNode
}

type PublicInclusionExclusionProof struct {
	Root utils.Hash  `cbor:"1,keyasint" json:"root"`
	Path [][2][]byte `cbor:"2,keyasint" json:"path"`
	Key  utils.Hash  `cbor:"3,keyasint" json:"key"`
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

		path[i][0] = proofNode.Path
		if len(proofNode.Data) > 0 {
			path[i][1] = proofNode.Data
		} else {
			path[i][1] = proofNode.Hash
		}
	}

	return &PublicInclusionExclusionProof{
		Root: proof.Root,
		Key:  proof.Key,
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
			Path: tuple[0],
		}

		if i == 0 {
			proofNode.Data = tuple[1]
		} else {
			proofNode.Hash = tuple[1]
		}

		path[i] = proofNode
	}

	return &InclusionExclusionProof{
		Root: proof.Root,
		Key:  proof.Key,
		Path: path,
	}
}

func (t *SMT) VerifyPublicProof(proof *PublicInclusionExclusionProof) (bool, error) {
	internalProof := proof.ToInternalProof()
	if internalProof == nil {
		return false, nil
	}

	if valid, err := t.VerifyProof(internalProof); err == nil && valid {
		return true, nil
	}

	if len(internalProof.Path) > 0 && internalProof.Path[0] != nil && len(internalProof.Path[0].Data) > 0 {
		internalProof.Path[0].Hash = internalProof.Path[0].Data
		internalProof.Path[0].Data = nil
		if valid, err := t.VerifyProof(internalProof); err == nil && valid {
			return true, nil
		}
	}

	return false, fmt.Errorf("public proof is neither a valid inclusion nor exclusion proof")
}

func (t *SMT) VerifyProof(proof *InclusionExclusionProof) (bool, error) {
	if proof == nil {
		return false, nil
	}

	if len(proof.Path) == 0 {
		return false, fmt.Errorf("proof path is empty")
	}

	if !bytes.Equal(proof.Root, t.GetRoot().GetHash()) {
		return false, fmt.Errorf("given proof document contains invalid root hash")
	}

	if len(proof.Path[0].Data) > 0 {
		if len(proof.Path) < 2 {
			return false, fmt.Errorf("proof path must contain at least leaf and parent witnesses")
		}

		recomputedRoot, err := t.recomputeRootFromProofBySpec(proof)
		if err == nil && bytes.Equal(recomputedRoot, proof.Root) {
			if err := verifyLeafPathMatchesProofKey(proof, t.HashAlgo); err == nil {
				return true, nil
			}
		}
	}

	if valid, err := t.verifyExclusionProof(proof); err == nil && valid {
		return true, nil
	}

	if len(proof.Path[0].Hash) > 0 {
		proof.Path[0].Data = proof.Path[0].Hash
		proof.Path[0].Hash = nil
		defer func() {
			proof.Path[0].Hash = proof.Path[0].Data
			proof.Path[0].Data = nil
		}()

		if len(proof.Path) < 2 {
			return false, fmt.Errorf("proof path must contain at least leaf and parent witnesses")
		}

		recomputedRoot, err := t.recomputeRootFromProofBySpec(proof)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(recomputedRoot, proof.Root) {
			return false, fmt.Errorf("calculated root hash does not match provided root hash")
		}
		if err := verifyLeafPathMatchesProofKey(proof, t.HashAlgo); err != nil {
			return false, err
		}
		return true, nil
	}

	return t.verifyExclusionProof(proof)
}

func (t *SMT) verifyExclusionProof(proof *InclusionExclusionProof) (bool, error) {
	if len(proof.Path) == 0 {
		return false, fmt.Errorf("proof path is empty")
	}

	fullPath, err := buildFullLeafPathFromProof(proof)
	if err != nil {
		return false, err
	}

	if proof.Path[0] == nil {
		return false, fmt.Errorf("proof contains nil node at index 0")
	}

	keyBitLen := utils.GetHashAlgoOutputBitCount(t.HashAlgo)
	if len(proof.Key) != keyBitLen/8 {
		return false, fmt.Errorf("proof key length %d does not match expected %d", len(proof.Key), keyBitLen/8)
	}

	validateDivergence := func() error {
		witnessPath, ok := decodeEncodedPath(proof.Path[0].Path)
		if !ok {
			return fmt.Errorf("proof path[0] is not a valid encoded path")
		}
		witnessBitLen := pathBitLen(witnessPath)
		parentBitLen := pathBitLen(fullPath) - witnessBitLen
		if witnessBitLen == 0 || parentBitLen < 0 {
			return fmt.Errorf("exclusion proof witness path is invalid")
		}
		if !pathMatchesKey(fullPath, proof.Key, parentBitLen) {
			return fmt.Errorf("exclusion proof parent path does not match proof key")
		}
		commonPrefixLen := pathCommonPrefixLenAtKey(witnessPath, 0, proof.Key, parentBitLen, keyBitLen)
		if commonPrefixLen == 0 {
			return fmt.Errorf("exclusion proof witness is not on the proof key branch")
		}
		if commonPrefixLen >= witnessBitLen {
			return fmt.Errorf("exclusion proof witness path does not diverge from proof key")
		}
		return nil
	}

	switch {
	case len(proof.Path[0].Data) > 0:
		if pathMatchesKey(fullPath, proof.Key, keyBitLen) {
			return false, fmt.Errorf("exclusion proof contains target leaf")
		}
		if err := validateDivergence(); err != nil {
			return false, err
		}
	case len(proof.Path[0].Hash) == 0:
		if !pathMatchesKey(fullPath, proof.Key, pathBitLen(fullPath)) {
			return false, fmt.Errorf("empty exclusion proof path does not match proof key")
		}
	default:
		if err := validateDivergence(); err != nil {
			return false, err
		}
	}

	recomputedRoot, err := t.recomputeRootFromProofBySpec(proof)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(recomputedRoot, proof.Root) {
		return false, fmt.Errorf("calculated root hash does not match provided root hash")
	}

	return true, nil
}

func appendPathSegment(dst *Path, segment *Path) {
	if dst == nil || segment == nil || segment.bitLen <= 0 {
		return
	}

	oldBitLen := dst.bitLen
	newBitLen := oldBitLen + segment.bitLen
	oldByteLen := pathByteLen(oldBitLen)
	newByteLen := pathByteLen(newBitLen)

	if cap(dst.bits) < newByteLen {
		grown := make([]byte, newByteLen)
		copy(grown, dst.bits[:oldByteLen])
		dst.bits = grown
	} else {
		dst.bits = dst.bits[:newByteLen]
		if newByteLen > oldByteLen {
			clear(dst.bits[oldByteLen:newByteLen])
		}
	}

	shift := oldBitLen & 7
	dstByteOffset := oldBitLen >> 3
	srcByteLen := pathByteLen(segment.bitLen)

	if shift == 0 {
		copy(dst.bits[dstByteOffset:], segment.bits[:srcByteLen])
		dst.bitLen = newBitLen
		return
	}

	for i := 0; i < srcByteLen; i++ {
		b := segment.bits[i]
		dstIdx := dstByteOffset + i
		dst.bits[dstIdx] |= b << shift
		if dstIdx+1 < len(dst.bits) {
			dst.bits[dstIdx+1] |= b >> (8 - shift)
		}
	}

	dst.bitLen = newBitLen
	maskUnusedPathBits(dst.bits, dst.bitLen)
}

func buildFullLeafPathFromProof(proof *InclusionExclusionProof) (*Path, error) {
	if proof == nil || len(proof.Path) == 0 {
		return nil, fmt.Errorf("cannot build leaf path from empty proof")
	}

	path := &Path{}

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

		appendPathSegment(path, segment)
	}

	return path, nil
}

func verifyLeafPathMatchesProofKey(proof *InclusionExclusionProof, hashAlgo utils.HashAlgo) error {
	fullPath, err := buildFullLeafPathFromProof(proof)
	if err != nil {
		return err
	}

	keyBitLen := utils.GetHashAlgoOutputBitCount(hashAlgo)
	if pathBitLen(fullPath) != keyBitLen {
		return fmt.Errorf("calculated leaf path bit length %d does not match key bit length %d", pathBitLen(fullPath), keyBitLen)
	}
	if len(proof.Key) != keyBitLen/8 {
		return fmt.Errorf("proof key length %d does not match expected %d", len(proof.Key), keyBitLen/8)
	}

	for bitIdx := 0; bitIdx < keyBitLen; bitIdx++ {
		if pathBit(fullPath, bitIdx) != keyPathBit(proof.Key, bitIdx) {
			return fmt.Errorf("calculated leaf path does not match proof key")
		}
	}

	return nil
}

func pathMatchesKey(path *Path, key []byte, bitLen int) bool {
	if path == nil {
		return bitLen == 0
	}
	if bitLen < 0 || bitLen > pathBitLen(path) {
		return false
	}
	if bitLen > len(key)*8 {
		return false
	}

	for bitIdx := 0; bitIdx < bitLen; bitIdx++ {
		if pathBit(path, bitIdx) != keyPathBit(key, bitIdx) {
			return false
		}
	}

	return true
}

func (t *SMT) recomputeRootFromProofBySpec(proof *InclusionExclusionProof) (utils.Hash, error) {
	if proof == nil || len(proof.Path) == 0 {
		return nil, fmt.Errorf("cannot recompute root from empty proof")
	}
	if proof.Path[0] == nil {
		return nil, fmt.Errorf("proof contains nil node at index 0")
	}

	var currentHash utils.Hash
	if len(proof.Path[0].Data) > 0 {
		currentHash = t.calculateLeafNodeHash(proof.Path[0].Path, proof.Path[0].Data, nil)
	} else {
		currentHash = proof.Path[0].Hash
	}

	for i := 1; i < len(proof.Path); i++ {
		if proof.Path[i] == nil {
			return nil, fmt.Errorf("proof contains nil node at index %d", i)
		}

		firstBit, ok := firstEncodedPathBit(proof.Path[i-1].Path)
		if !ok {
			return nil, fmt.Errorf("proof path[%d] has no meaningful bits", i-1)
		}

		if firstBit {
			currentHash = t.calculateBranchNodeHash(proof.Path[i].Path, proof.Path[i].Hash, currentHash, nil)
		} else {
			currentHash = t.calculateBranchNodeHash(proof.Path[i].Path, currentHash, proof.Path[i].Hash, nil)
		}
	}

	return currentHash, nil
}

func firstEncodedPathBit(encodedPath []byte) (bool, bool) {
	if len(encodedPath) == 0 {
		return false, false
	}

	bitLen, n := binary.Uvarint(encodedPath)
	if n <= 0 || bitLen == 0 || len(encodedPath) < n+1 {
		return false, false
	}

	return (encodedPath[n] & 1) != 0, true
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

func (node *Node) setPath(path *Path) {
	if node == nil {
		return
	}

	node.Path = path
	node.encodedPath = encodePath(path)
}

func (node *Node) hashDst() utils.Hash {
	if node == nil || cap(node.Hash) == 0 {
		return nil
	}
	return node.Hash[:0]
}

func (node *Node) encodedPathForProof() []byte {
	if node == nil {
		return nil
	}

	return node.encodedPath
}

func (node *Node) CalculateLeafHash(tree *SMT) {
	if node == nil {
		return
	}

	node.Hash = tree.calculateLeafNodeHash(node.encodedPath, node.Data, node.hashDst())
}

func (node *Node) CalculateBranchHash(tree *SMT) {
	if node == nil {
		return
	}

	node.Hash = tree.calculateBranchNodeHash(node.encodedPath, node.GetLeftNode().GetHash(), node.GetRightNode().GetHash(), node.hashDst())
}

func NewSMT(hashAlgo utils.HashAlgo, appendOnly bool) *SMT {
	smt := &SMT{
		HashAlgo:   hashAlgo,
		AppendOnly: appendOnly,
		dataHasher: utils.NewDataHasher(hashAlgo),
	}
	smt.Init()
	return smt
}

func newBranchNode(path *Path) *Node {
	node := &Node{
		Data:      nil,
		LeftNode:  nil,
		RightNode: nil,
		IsLeaf:    false,
		Key:       nil,
	}
	node.setPath(path)
	return node
}

func (t *SMT) Init() {
	// set initial values
	rootNode := newBranchNode(nil)
	rootNode.Hash = t.calculateBranchNodeHash(nil, nil, nil, rootNode.hashDst())
	t.Root = rootNode
}

func (t *SMT) GetRoot() *Node {
	if t.Root == nil {
		rootNode := newBranchNode(nil)
		rootNode.Hash = t.calculateBranchNodeHash(nil, nil, nil, rootNode.hashDst())
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

// InsertHashed inserts data using an already-hashed key without hashing it again.
func (t *SMT) InsertHashed(keyHash utils.Hash, data []byte) (bool, error) {
	chosenPath := chooseNewPath(keyHash, nil)
	if pathBitLen(chosenPath) <= 0 {
		return false, fmt.Errorf("cannot insert with empty path")
	}

	_, err := t.insertPrepared(t.GetRoot(), chosenPath, 0, keyHash, data, t.AppendOnly)
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

// ChooseNewKey returns both the key and the number of meaningful bits for correct comparison.
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
	return newPath
}

func newLeafNode(tree *SMT, keyHash utils.Hash, data []byte, path *Path) *Node {
	node := &Node{
		Data:      data,
		LeftNode:  nil,
		RightNode: nil,
		IsLeaf:    true,
		Key:       keyHash,
	}
	node.setPath(path)
	node.CalculateLeafHash(tree)
	return node
}

func recalculateNodeHash(node *Node, tree *SMT) {
	if node == nil {
		return
	}
	if node.IsLeaf {
		node.CalculateLeafHash(tree)
		return
	}
	node.CalculateBranchHash(tree)
}

func (t *SMT) insert(key []byte, data []byte, currentRoot *Node, newPath *Path, appendOnly bool) (*Node, error) {
	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
	nodeData := data

	chosenPath := chooseNewPath(nodeKeyHash, newPath)
	if pathBitLen(chosenPath) <= 0 {
		return nil, fmt.Errorf("cannot insert with empty path")
	}

	return t.insertPrepared(currentRoot, chosenPath, 0, nodeKeyHash, nodeData, appendOnly)
}

func (t *SMT) insertPrepared(currentRoot *Node, chosenPath *Path, pathOffset int, nodeKeyHash utils.Hash, nodeData []byte, appendOnly bool) (*Node, error) {
	if pathOffset < 0 {
		pathOffset = 0
	}
	if pathOffset >= pathBitLen(chosenPath) {
		return nil, fmt.Errorf("cannot insert with exhausted path")
	}

	goLeft := pathBit(chosenPath, pathOffset) == 0

	if goLeft {
		if currentRoot.LeftNode == nil {
			currentRoot.LeftNode = newLeafNode(t, nodeKeyHash, nodeData, pathCutPrefix(chosenPath, pathOffset))
			currentRoot.CalculateBranchHash(t)
			return currentRoot, nil
		}

		if currentRoot.LeftNode.IsLeaf {
			if bytes.Equal(currentRoot.LeftNode.Key, nodeKeyHash) {
				if appendOnly {
					return nil, fmt.Errorf("duplicate key insert in append-only mode")
				}
				currentRoot.LeftNode.Data = nodeData
				currentRoot.LeftNode.CalculateLeafHash(t)
				currentRoot.CalculateBranchHash(t)
				return currentRoot, nil
			}

			existingPath := currentRoot.LeftNode.Path
			commonPrefixLen := pathCommonPrefixLenAt(existingPath, 0, chosenPath, pathOffset)
			commonPrefix := pathSlice(chosenPath, pathOffset, commonPrefixLen)

			branchNode := newBranchNode(commonPrefix)

			newNodePath := pathCutPrefix(chosenPath, pathOffset+commonPrefixLen)
			oldNodePath := pathCutPrefix(existingPath, commonPrefixLen)

			currentRoot.LeftNode.setPath(oldNodePath)
			currentRoot.LeftNode.CalculateLeafHash(t)

			if pathBit(existingPath, commonPrefixLen) == 1 {
				branchNode.RightNode = currentRoot.LeftNode
				branchNode.LeftNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
			} else {
				branchNode.LeftNode = currentRoot.LeftNode
				branchNode.RightNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
			}

			branchNode.CalculateBranchHash(t)
			currentRoot.LeftNode = branchNode
			currentRoot.CalculateBranchHash(t)
			return currentRoot, nil
		}

		existingPath := currentRoot.LeftNode.Path
		commonPrefixLen := pathCommonPrefixLenAt(existingPath, 0, chosenPath, pathOffset)

		if commonPrefixLen == pathBitLen(currentRoot.LeftNode.Path) {
			if _, err := t.insertPrepared(currentRoot.LeftNode, chosenPath, pathOffset+commonPrefixLen, nodeKeyHash, nodeData, appendOnly); err != nil {
				return nil, err
			}
			currentRoot.CalculateBranchHash(t)
			return currentRoot, nil
		}

		commonPrefix := pathSlice(chosenPath, pathOffset, commonPrefixLen)
		newBranchNode := newBranchNode(commonPrefix)

		newNodePath := pathCutPrefix(chosenPath, pathOffset+commonPrefixLen)
		oldBranchPath := pathCutPrefix(existingPath, commonPrefixLen)

		currentRoot.LeftNode.setPath(oldBranchPath)
		recalculateNodeHash(currentRoot.LeftNode, t)

		if pathBit(existingPath, commonPrefixLen) == 1 {
			newBranchNode.RightNode = currentRoot.LeftNode
			newBranchNode.LeftNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
		} else {
			newBranchNode.LeftNode = currentRoot.LeftNode
			newBranchNode.RightNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
		}

		newBranchNode.CalculateBranchHash(t)
		currentRoot.LeftNode = newBranchNode
		currentRoot.CalculateBranchHash(t)
		return currentRoot, nil
	}

	if currentRoot.RightNode == nil {
		currentRoot.RightNode = newLeafNode(t, nodeKeyHash, nodeData, pathCutPrefix(chosenPath, pathOffset))
		currentRoot.CalculateBranchHash(t)
		return currentRoot, nil
	}

	if currentRoot.RightNode.IsLeaf {
		if bytes.Equal(currentRoot.RightNode.Key, nodeKeyHash) {
			if appendOnly {
				return nil, fmt.Errorf("duplicate key insert in append-only mode")
			}
			currentRoot.RightNode.Data = nodeData
			currentRoot.RightNode.CalculateLeafHash(t)
			currentRoot.CalculateBranchHash(t)
			return currentRoot, nil
		}

		existingPath := currentRoot.RightNode.Path
		commonPrefixLen := pathCommonPrefixLenAt(existingPath, 0, chosenPath, pathOffset)
		commonPrefix := pathSlice(chosenPath, pathOffset, commonPrefixLen)

		branchNode := newBranchNode(commonPrefix)

		newNodePath := pathCutPrefix(chosenPath, pathOffset+commonPrefixLen)
		oldNodePath := pathCutPrefix(existingPath, commonPrefixLen)

		currentRoot.RightNode.setPath(oldNodePath)
		currentRoot.RightNode.CalculateLeafHash(t)

		if pathBit(existingPath, commonPrefixLen) == 1 {
			branchNode.RightNode = currentRoot.RightNode
			branchNode.LeftNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
		} else {
			branchNode.LeftNode = currentRoot.RightNode
			branchNode.RightNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
		}

		branchNode.CalculateBranchHash(t)
		currentRoot.RightNode = branchNode
		currentRoot.CalculateBranchHash(t)
		return currentRoot, nil
	}

	existingPath := currentRoot.RightNode.Path
	commonPrefixLen := pathCommonPrefixLenAt(existingPath, 0, chosenPath, pathOffset)

	if commonPrefixLen == pathBitLen(currentRoot.RightNode.Path) {
		if _, err := t.insertPrepared(currentRoot.RightNode, chosenPath, pathOffset+commonPrefixLen, nodeKeyHash, nodeData, appendOnly); err != nil {
			return nil, err
		}
		currentRoot.CalculateBranchHash(t)
		return currentRoot, nil
	}

	commonPrefix := pathSlice(chosenPath, pathOffset, commonPrefixLen)
	newBranchNode := newBranchNode(commonPrefix)

	newNodePath := pathCutPrefix(chosenPath, pathOffset+commonPrefixLen)
	oldBranchPath := pathCutPrefix(existingPath, commonPrefixLen)

	currentRoot.RightNode.setPath(oldBranchPath)
	recalculateNodeHash(currentRoot.RightNode, t)

	if pathBit(existingPath, commonPrefixLen) == 1 {
		newBranchNode.RightNode = currentRoot.RightNode
		newBranchNode.LeftNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
	} else {
		newBranchNode.LeftNode = currentRoot.RightNode
		newBranchNode.RightNode = newLeafNode(t, nodeKeyHash, nodeData, newNodePath)
	}

	newBranchNode.CalculateBranchHash(t)
	currentRoot.RightNode = newBranchNode
	currentRoot.CalculateBranchHash(t)

	return currentRoot, nil
}

func (t *SMT) GenerateInclusionExclusionProof(key []byte) (*InclusionExclusionProof, error) {
	rootNode := t.GetRoot()
	if rootNode.LeftNode == nil && rootNode.RightNode == nil {
		return nil, fmt.Errorf("Cannot generate inclusion proof for an empty tree")
	}

	keyBitSize := utils.GetHashAlgoOutputBitCount(t.HashAlgo)
	proof := &InclusionExclusionProof{
		Root: rootNode.GetHash(),
		Key:  key,
		Path: make([]*ProofNode, 0, keyBitSize+1),
	}

	i := 0
	currNode := rootNode
	pathMismatch := false

	for i < keyBitSize && currNode != nil && !currNode.IsLeaf {
		goLeft := keyPathBit(key, i) == 0

		proofNode := &ProofNode{
			Path: currNode.encodedPathForProof(),
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
			proof.Path = append(proof.Path, &ProofNode{
				Path: pathSlice(pathFromKeyBytes(key, keyBitSize), i, 1).Encode(),
				Hash: nil,
				Data: nil,
			})
			currNode = nil
			break
		}

		commonPrefixLen := pathCommonPrefixLenAtKey(nextNode.Path, 0, key, i, keyBitSize)
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
				Path: currNode.encodedPathForProof(),
				Hash: nil,
				Data: currNode.Data,
			})
		} else if pathMismatch || i >= keyBitSize {
			proof.Path = append(proof.Path, &ProofNode{
				Path: currNode.encodedPathForProof(),
				Hash: currNode.Hash,
				Data: nil,
			})
		}
	}

	slices.Reverse(proof.Path)
	return proof, nil
}
