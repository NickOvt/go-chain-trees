package smt

import (
	"encoding/hex"

	"github.com/NickOvt/go-chain-trees/utils"
)

func EncodePath(data []byte, depth int) []byte {
	if len(data) == 0 {
		return data
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

	return final
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
	Root           utils.Hash
	Nodes          map[string]*Node
	emptyHashCache map[int]utils.Hash
	emptyHash      utils.Hash
}

type Node struct {
	Key    utils.Hash     // in case of a non-leaf node it will be nil
	Data   utils.CBORData // in case of a non-leaf node it will be nil
	Hash   utils.Hash     // present on every node
	Path   []byte         // nil in case of a root node
	IsLeaf bool
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
	t.Root = t.GetEmptyHash()
	t.Nodes = make(map[string]*Node)
	t.emptyHashCache = make(map[int]utils.Hash)

	t.BuildHashCache(t.HashAlgo)
}

// Update insert or update value
//func (t *SMT) Update(key []byte, data []byte) (bool, error) {
//	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
//	nodePath := nodeKeyHash
//
//	if existingNode, ok := t.Nodes[hex.EncodeToString(nodePath)]; ok {
//		// have duplicate, update data
//		cborData, err := utils.EncodeCBOR(data)
//
//		if err != nil {
//			return false, err
//		}
//
//		existingNode.Data = cborData
//
//		// TODO: Update parents
//	}
//
//	// node does not exist, insert
//	node := &Node{Key: nodeKeyHash, Data: data, Path: nodePath}
//	t.Nodes[hex.EncodeToString(nodePath)] = node
//
//	// in SMT same level sibling has last bit flipped
//	siblingPath := utils.FlipLastBit(nodePath)
//
//	if _, ok := t.Nodes[hex.EncodeToString(siblingPath)]; ok { // existingSibling, ok
//		// sibling exists -> find parent
//
//		parentPath := EncodePath(nodePath, 1) // parent path is the intersection of two children nodes, that is path without last bit
//
//		if _, ok := t.Nodes[hex.EncodeToString(parentPath)]; ok {
//			// parent found
//			// TODO: update parents
//		}
//
//		// parent not found -> insert parent
//		parentNode := &Node{Key: parentPath, Path: parentPath}
//		t.Nodes[hex.EncodeToString(parentPath)] = parentNode
//
//		// TODO: update parents
//	} else {
//		// current level sibling does not exist
//	}
//
//	return true, nil
//}

// Update insert or update value
func (t *SMT) Update(key []byte, data []byte) (bool, error) {
	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
	nodePath := nodeKeyHash[:]

	newRoot, err := t.updateRecursive(nodePath, data, 0, t.Root)
	if err != nil {
		return false, err
	}

	t.Root = newRoot
	return true, nil

}

func (t *SMT) updateRecursive(nodePath utils.Hash, data []byte, depth int, currentRoot utils.Hash) (utils.Hash, error) {
	if depth == utils.GetHashAlgoOutputBitCount(t.HashAlgo) {
		nodePathEnc := EncodePath(nodePath, depth)

		leaf := &Node{
			Key:    nodePathEnc,
			Data:   data,
			Path:   nodePathEnc,
			IsLeaf: true,
		}

		t.Nodes[hex.EncodeToString(nodePathEnc)] = leaf
		return leaf.Key, nil
	}

	if hex.EncodeToString(currentRoot) == hex.EncodeToString(t.GetEmptyHash()) {
		// root is empty, first leaf inserted

		nodePathEnc := EncodePath(nodePath, depth)

		leaf := &Node{
			Key:    nodePathEnc,
			Data:   data,
			Path:   nodePathEnc,
			IsLeaf: true,
		}

		t.Nodes[hex.EncodeToString(nodePathEnc)] = leaf
		return leaf.Key, nil
	}

	existingNode, _ := t.Nodes[hex.EncodeToString(currentRoot)]

	return []byte{}, nil
}
