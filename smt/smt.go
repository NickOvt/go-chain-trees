package smt

import (
	"github.com/NickOvt/go-chain-trees/utils"
)

func EncodePath(data []byte, depth int) ([]byte, int) {
	if len(data) == 0 {
		return []byte{byte(1)}, 0
	}

	totalBits := depth
	if totalBits > len(data)*8 {
		totalBits = len(data) * 8
	}

	result := make([]byte, (totalBits+7)/8) // round up for padding

	for i := 0; i < totalBits; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)

		if data[byteIdx]&(1<<bitIdx) != 0 { // original data had 1 at this position
			result[byteIdx] |= 1 << bitIdx // set non 0 bits to 1 in result
		}
	}

	// prepend 1-bit
	paddingBits := (8 - (totalBits+1)%8) % 8
	finalLen := (totalBits + 1 + paddingBits + 7) / 8
	final := make([]byte, finalLen)

	// set 1-bit marker
	final[0] = 1 << (7 - paddingBits)

	// copy data after 1-bit
	for i := 0; i < totalBits; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)

		destBitPos := paddingBits + 1 + i  // shift by padding
		destByteIdx := destBitPos / 8      // shift by padding
		destBitIdx := 7 - (destBitPos % 8) // shift by padding

		if result[byteIdx]&(1<<bitIdx) != 0 {
			final[destByteIdx] |= 1 << destBitIdx
		}
	}

	return final, totalBits
}

func DecodePath(encoded []byte) []byte {
	if len(encoded) == 0 {
		return encoded
	}

	// find the 1-bit marker
	var markerPos int
	found := false
	for i := 0; i < 8; i++ { // check only first byte
		if encoded[0]&(1<<(7-i)) != 0 {
			markerPos = i
			found = true
			break
		}
	}

	if !found {
		return encoded
	}

	// extract bits after marker
	dataBits := len(encoded)*8 - markerPos - 1
	if dataBits <= 0 {
		return encoded
	}

	result := make([]byte, (dataBits+7)/8)

	for i := 0; i < dataBits; i++ {
		srcBitPos := markerPos + 1 + i
		srcByteIdx := srcBitPos / 8
		srcBitIdx := 7 - (srcBitPos % 8)

		destByteIdx := i / 8
		destBitIdx := 7 - (i % 8)

		if encoded[srcByteIdx]&(1<<srcBitIdx) != 0 {
			result[destByteIdx] |= 1 << destBitIdx
		}
	}

	return result
}

type SMT struct {
	HashAlgo       utils.HashAlgo
	Root           *Node
	emptyHashCache map[int]utils.Hash
	emptyHash      utils.Hash
}

type Node struct {
	Data      utils.CBORData // in case of a non-leaf node it will be nil
	Hash      utils.Hash     // present on every node. hash(path, data)
	Path      []byte         // nil in case of a root node. Encoded path from parent to this node
	LeftHash  utils.Hash     // For branch nodes
	RightHash utils.Hash     // For branch nodes
	IsLeaf    bool
}

func NewSMT(hashAlgo utils.HashAlgo) *SMT {
	smt := &SMT{HashAlgo: hashAlgo}
	smt.Init()
	return smt
}

func (t *SMT) GetEmptyHash() utils.Hash {
	if t.emptyHash == nil {
		t.emptyHash = utils.GenerateNullHash(t.HashAlgo)
	}
	return t.emptyHash
}

func (t *SMT) GetEmptyHashForLevel(hashAlgo utils.HashAlgo, level int) utils.Hash {
	if cachedEmptyHash, ok := t.emptyHashCache[level]; ok {
		return cachedEmptyHash
	}

	var hash utils.Hash
	if level == 0 {
		hash = t.GetEmptyHash()
	} else {
		prevHash := t.GetEmptyHashForLevel(hashAlgo, level-1)
		hash = utils.ConcatHashesAndGenerateHash(hashAlgo, prevHash, prevHash)
	}

	t.emptyHashCache[level] = hash
	return hash
}

func (t *SMT) BuildHashCache(hashAlgo utils.HashAlgo) {
	hashAlgoBitCount := utils.GetHashAlgoOutputBitCount(hashAlgo)

	// Calculate and fill hash
	for level := range hashAlgoBitCount {
		t.GetEmptyHashForLevel(hashAlgo, level)
	}
}

func (t *SMT) Init() {
	if t.emptyHash == nil {
		t.emptyHash = utils.GenerateNullHash(t.HashAlgo)
	}

	// set initial values

	rootNode := &Node{
		Data:      nil,
		LeftHash:  nil,
		RightHash: nil,
		Hash:      utils.ConcatHashesAndGenerateHash(t.HashAlgo, nil, nil, nil),
		IsLeaf:    false,
	}
	t.Root = rootNode

	t.emptyHashCache = make(map[int]utils.Hash)

	t.BuildHashCache(t.HashAlgo)
}
