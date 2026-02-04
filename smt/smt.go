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
	Key       utils.Hash     // Key of node (will be hashed by the tree hashAlgo), nil for branch node
	Data      utils.CBORData // in case of a non-leaf node it will be nil
	Hash      utils.Hash     // present on every node. hash(path, data), for branch nodes hash(path, leftHash, rightHash)
	Path      []byte         // nil in case of a root node. Encoded path from parent to this node
	LeftNode  *Node          // For branch nodes
	RightNode *Node          // For branch nodes
	IsLeaf    bool
}

// TODO: add methods to safely get LeftNode and RightNode and to calculate Hash

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
		LeftNode:  nil,
		RightNode: nil,
		Path:      nil,
		Hash:      utils.ConcatHashesAndGenerateHash(t.HashAlgo, nil, nil, nil),
		IsLeaf:    false,
	}
	t.Root = rootNode

	t.emptyHashCache = make(map[int]utils.Hash)

	t.BuildHashCache(t.HashAlgo)
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
	_, err := t.insert(key, data, t.GetRoot())

	if err != nil {
		return false, err
	}

	return true, nil
}

func (t *SMT) insert(key []byte, data []byte, currentRoot *Node) (*Node, error) {
	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
	nodeDataCbor, err := utils.EncodeCBOR(data)

	if err != nil {
		return nil, err
	}

	goLeft := utils.GetBit(nodeKeyHash, 0)

	if goLeft {
		// check left of current root
		if currentRoot.LeftNode == nil {
			// no node at all, insert directly new leaf
			path := utils.ReverseBits(nodeKeyHash) // path to parent is reversed key
			currentRoot.LeftNode = &Node{
				Data:      nodeDataCbor,
				LeftNode:  nil,
				RightNode: nil,
				Path:      path,
				Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, path, nodeDataCbor),
				IsLeaf:    true,
				Key:       nodeKeyHash,
			}
			// TODO: recalculate root's hash
		} else {
			// have node at left side
			if currentRoot.LeftNode.IsLeaf {
				// left is leaf, create branch
				// find common prefix of inserted and current leaf on the left
				commonPrefix, commonPrefixLen, commonPrefixPaddingLen := utils.FindCommonBitPrefix(currentRoot.LeftNode.Key, nodeKeyHash)
				branchNodePath, _ := EncodePath(utils.ReverseBits(commonPrefix), commonPrefixPaddingLen) // use padding here as we have reversed prefix
				branchNode := &Node{
					Data:      nil,
					LeftNode:  nil,
					RightNode: nil,
					Path:      branchNodePath,
					IsLeaf:    false,
					Key:       nil,
				}

				newNodeKeyCut, _, newKeyPaddingLen := utils.RemoveFirstNBits(nodeKeyHash, commonPrefixLen)
				newNodePath, _ := EncodePath(utils.ReverseBits(newNodeKeyCut), newKeyPaddingLen)

				oldNodeKeyCut, _, oldKeyPaddingLen := utils.RemoveFirstNBits(currentRoot.LeftNode.Key, commonPrefixLen)
				oldNodeNewPath, _ := EncodePath(utils.ReverseBits(oldNodeKeyCut), oldKeyPaddingLen)
				currentRoot.LeftNode.Path = oldNodeNewPath
				currentRoot.LeftNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, currentRoot.LeftNode.Path, currentRoot.LeftNode.Data)

				if utils.GetBit(currentRoot.LeftNode.Key, commonPrefixLen) {
					// old left node goes right now, new goes left
					branchNode.RightNode = currentRoot.LeftNode
					branchNode.LeftNode = &Node{
						Data:      nodeDataCbor,
						LeftNode:  nil,
						RightNode: nil,
						Path:      newNodePath,
						IsLeaf:    true,
						Key:       nodeKeyHash,
						Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
					} // the inserted node goes here
				} else {
					// old left node stays left, new goes right
					branchNode.RightNode = &Node{
						Data:      nodeDataCbor,
						LeftNode:  nil,
						RightNode: nil,
						Path:      newNodePath,
						IsLeaf:    true,
						Key:       nodeKeyHash,
						Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
					} // the inserted node goes here
					branchNode.LeftNode = currentRoot.LeftNode
				}
				branchNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, branchNode.Path, branchNode.LeftNode.Hash, branchNode.RightNode.Hash)
				currentRoot.LeftNode = branchNode
				// TODO: recalculate root's hash
			} else {
				// left is branch, recurse
			}
		}
	} else {
		// check right of current root
		if currentRoot.RightNode == nil {
			// no node at all
			path := utils.ReverseBits(nodeKeyHash) // path to parent is reversed key
			currentRoot.RightNode = &Node{
				Data:      nodeDataCbor,
				LeftNode:  nil,
				RightNode: nil,
				Path:      path,
				Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, path, nodeDataCbor),
				IsLeaf:    true,
				Key:       nodeKeyHash,
			}
			// TODO: recalculate root's hash
		} else {
			// have node at right side
			if currentRoot.RightNode.IsLeaf {
				// right is leaf, create branch
				// find common prefix of inserted and current leaf on the left
				commonPrefix, commonPrefixLen, commonPrefixPaddingLen := utils.FindCommonBitPrefix(currentRoot.RightNode.Key, nodeKeyHash)
				branchNodePath, _ := EncodePath(utils.ReverseBits(commonPrefix), commonPrefixPaddingLen) // use padding here as we have reversed prefix
				branchNode := &Node{
					Data:      nil,
					LeftNode:  nil,
					RightNode: nil,
					Path:      branchNodePath,
					IsLeaf:    false,
					Key:       nil,
				}

				newNodeKeyCut, _, newKeyPaddingLen := utils.RemoveFirstNBits(nodeKeyHash, commonPrefixLen)
				newNodePath, _ := EncodePath(utils.ReverseBits(newNodeKeyCut), newKeyPaddingLen)

				oldNodeKeyCut, _, oldKeyPaddingLen := utils.RemoveFirstNBits(currentRoot.RightNode.Key, commonPrefixLen)
				oldNodeNewPath, _ := EncodePath(utils.ReverseBits(oldNodeKeyCut), oldKeyPaddingLen)
				currentRoot.RightNode.Path = oldNodeNewPath
				currentRoot.RightNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, currentRoot.RightNode.Path, currentRoot.LeftNode.Data)

				if utils.GetBit(currentRoot.RightNode.Key, commonPrefixLen) {
					// old right node stays right now, new goes left
					branchNode.RightNode = currentRoot.RightNode
					branchNode.LeftNode = &Node{
						Data:      nodeDataCbor,
						LeftNode:  nil,
						RightNode: nil,
						Path:      newNodePath,
						IsLeaf:    true,
						Key:       nodeKeyHash,
						Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
					} // the inserted node goes here
				} else {
					// old right node goes left, new goes right
					branchNode.RightNode = &Node{
						Data:      nodeDataCbor,
						LeftNode:  nil,
						RightNode: nil,
						Path:      newNodePath,
						IsLeaf:    true,
						Key:       nodeKeyHash,
						Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
					} // the inserted node goes here
					branchNode.LeftNode = currentRoot.RightNode
				}
				branchNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, branchNode.Path, branchNode.LeftNode.Hash, branchNode.RightNode.Hash)
				currentRoot.RightNode = branchNode
				// TODO: recalculate root's hash
			} else {
				// right is branch, recurse
			}
		}
	}

	return currentRoot, nil
}
