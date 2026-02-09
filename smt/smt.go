package smt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

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

// left aligned, padded with 0 in result
func DecodePath(encoded []byte) ([]byte, int) {
	if len(encoded) == 0 {
		return encoded, 0
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
		return encoded, 0
	}

	// extract bits after marker
	dataBits := len(encoded)*8 - markerPos - 1
	if dataBits <= 0 {
		return []byte{}, 0
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

	// Calculate trailing padding in the last byte
	trailingPadding := (8 - (dataBits % 8)) % 8

	return result, trailingPadding
}

// CalculateKeyFromPath calculates a key from an encoded path by decoding it,
// reversing the bits, and removing padding. This produces a key that can be
// correctly compared with other keys for prefix calculations.
//
// The function returns both the calculated key and the number of meaningful bits,
// which is essential for correct bit-level comparisons when the key doesn't fill
// complete bytes.
//
// Parameters:
//   - encodedPath: The encoded path from which to calculate the key
//
// Returns:
//   - []byte: The calculated key (left-aligned, zero-padded on the right)
//   - int: Number of meaningful bits in the key (excluding padding)
//
// Example:
//   - If encodedPath represents path "0" (1 bit), the result will be a byte array
//     with the first bit set to 0 and remaining bits as padding: [0b00000000]
//     with meaningfulBits = 1
func CalculateKeyFromPath(encodedPath []byte) ([]byte, int) {
	if encodedPath == nil || len(encodedPath) == 0 {
		return []byte{}, 0
	}

	// Decode the path to get the bit sequence and padding information
	decodedPath, trailingPadding := DecodePath(encodedPath)

	// Calculate meaningful bits (total bits minus padding)
	meaningfulBits := len(decodedPath)*8 - trailingPadding

	if meaningfulBits <= 0 {
		return []byte{}, 0
	}

	// Reverse bits to get the key representation
	// (paths are stored reversed relative to keys)
	reversedKey := utils.ReverseBits(decodedPath)

	// Remove the trailing padding that became leading padding after reversal
	// This ensures the key is left-aligned with meaningful bits at the start
	key, _, _ := utils.RemoveFirstNBits(reversedKey, trailingPadding)

	return key, meaningfulBits
}

// EncodeKeyBitsAsPath converts a key-oriented bitstring into the canonical encoded path form.
//
// keyBits are expected to be left-aligned with meaningfulBits indicating how many bits are valid.
// Paths are stored bit-reversed relative to keys, so this helper reverses and re-aligns bits
// before delegating to EncodePath.
func EncodeKeyBitsAsPath(keyBits []byte, meaningfulBits int) []byte {
	if meaningfulBits <= 0 {
		encoded, _ := EncodePath(nil, 0)
		return encoded
	}

	paddingBits := (8 - (meaningfulBits % 8)) % 8
	reversed := utils.ReverseBits(keyBits)
	reversedAligned, _, _ := utils.RemoveFirstNBits(reversed, paddingBits)
	encoded, _ := EncodePath(reversedAligned, meaningfulBits)

	return encoded
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
	return node.LeftNode
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

	node.Hash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, node.Path, node.Data)
}

func (node *Node) CalculateBranchHash(hashAlgo utils.HashAlgo) {
	if node == nil {
		return
	}

	node.Hash = utils.ConcatDataAndGenerateCombinedHash(hashAlgo, node.Path, node.GetLeftNode().GetHash(), node.GetRightNode().GetHash())
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
	_, err := t.insert(key, data, t.GetRoot(), nil)

	if err != nil {
		return false, err
	}

	return true, nil
}

// PrintTree writes a visual representation of the SMT to stdout.
// Branch nodes display their decoded raw path bitstrings, while leaf nodes
// display their hashes.
func (t *SMT) PrintTree() {
	t.WriteTree(os.Stdout)
}

// WriteTree writes a visual representation of the SMT to the provided writer.
// Branch nodes display their decoded raw path bitstrings, while leaf nodes
// display their hashes.
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

func formatBranchPath(encodedPath []byte) string {
	if len(encodedPath) == 0 {
		return "<root>"
	}

	bits := decodePathToRawBits(encodedPath)
	if bits == "" {
		return "<empty>"
	}

	return bits
}

func formatLeafPath(encodedPath []byte) string {
	bits := decodePathToRawBits(encodedPath)
	if bits == "" {
		return "<empty>"
	}

	return bits
}

func decodePathToRawBits(encodedPath []byte) string {
	decoded, trailingPadding := DecodePath(encodedPath)
	meaningfulBits := len(decoded)*8 - trailingPadding
	if meaningfulBits <= 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(meaningfulBits)
	for i := 0; i < meaningfulBits; i++ {
		if utils.GetBit(decoded, i) {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
	}

	return sb.String()
}

// ChooseNewKey Returns both the key and the number of meaningful bits for correct comparison.
func ChooseNewKey(originalKeyHash utils.Hash, newEncodedPath []byte) ([]byte, int) {
	if newEncodedPath == nil || len(newEncodedPath) == 0 {
		return originalKeyHash, len(originalKeyHash) * 8
	}

	return CalculateKeyFromPath(newEncodedPath)
}

// removeFirstNBitsWithLen removes the first n meaningful bits from a bitstring.
// Unlike utils.RemoveFirstNBits, it respects bitLen and ignores right-side byte padding.
func removeFirstNBitsWithLen(data []byte, bitLen int, n int) ([]byte, int) {
	if bitLen <= 0 || n >= bitLen {
		return []byte{}, 0
	}

	if n < 0 {
		n = 0
	}

	remainingBits := bitLen - n
	result := make([]byte, (remainingBits+7)/8)

	for i := 0; i < remainingBits; i++ {
		if utils.GetBit(data, n+i) {
			result[i/8] |= 1 << (7 - (i % 8))
		}
	}

	return result, remainingBits
}

func (t *SMT) insert(key []byte, data []byte, currentRoot *Node, newEncodedPath []byte) (*Node, error) {
	nodeKeyHash := utils.GenerateHash(t.HashAlgo, key)
	nodeDataCbor, err := utils.EncodeCBOR(data)

	if err != nil {
		return nil, err
	}

	chosenKey, chosenKeyBitLen := ChooseNewKey(nodeKeyHash, newEncodedPath)
	goLeft := !utils.GetBit(chosenKey, 0)

	if goLeft {
		// check left of current root
		if currentRoot.LeftNode == nil {
			// no node at all, insert directly new leaf
			var path []byte
			if newEncodedPath == nil {
				// not in recursion
				path = EncodeKeyBitsAsPath(nodeKeyHash, len(nodeKeyHash)*8)
			} else {
				// in recursion
				path = newEncodedPath
			}

			currentRoot.LeftNode = &Node{
				Data:      nodeDataCbor,
				LeftNode:  nil,
				RightNode: nil,
				Path:      path,
				Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, path, nodeDataCbor),
				IsLeaf:    true,
				Key:       nodeKeyHash,
			}
			currentRoot.CalculateBranchHash(t.HashAlgo)
		} else {
			// have node at left side
			if currentRoot.LeftNode.IsLeaf {
				// left is leaf, create branch
				// find common prefix of inserted and current leaf on the left
				leftKey, leftMeaningfulBits := CalculateKeyFromPath(currentRoot.LeftNode.Path)
				commonPrefix, commonPrefixLen, _ := utils.FindCommonBitPrefixWithLen(leftKey, leftMeaningfulBits, chosenKey, chosenKeyBitLen)
				branchNodePath := EncodeKeyBitsAsPath(commonPrefix, commonPrefixLen)
				branchNode := &Node{
					Data:      nil,
					LeftNode:  nil,
					RightNode: nil,
					Path:      branchNodePath,
					IsLeaf:    false,
					Key:       nil,
				}

				newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
				newNodePath := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

				oldNodeKeyCut, oldNodeKeyBitLen := removeFirstNBitsWithLen(leftKey, leftMeaningfulBits, commonPrefixLen)
				oldNodeNewPath := EncodeKeyBitsAsPath(oldNodeKeyCut, oldNodeKeyBitLen)
				currentRoot.LeftNode.Path = oldNodeNewPath
				currentRoot.LeftNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, currentRoot.LeftNode.Path, currentRoot.LeftNode.Data)

				if utils.GetBit(leftKey, commonPrefixLen) {
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
				branchNode.CalculateBranchHash(t.HashAlgo)
				currentRoot.LeftNode = branchNode
				currentRoot.CalculateBranchHash(t.HashAlgo)
			} else {
				// left is branch
				// 1. find common prefix of inserting node and existing branch node
				leftKey, leftMeaningfulBits := CalculateKeyFromPath(currentRoot.LeftNode.Path)
				commonPrefix, commonPrefixLen, _ := utils.FindCommonBitPrefixWithLen(leftKey, leftMeaningfulBits, chosenKey, chosenKeyBitLen)
				// 2. check if path from prefix is equal to branch's path
				branchNodePath := EncodeKeyBitsAsPath(commonPrefix, commonPrefixLen)
				if bytes.Equal(branchNodePath, currentRoot.LeftNode.Path) {
					// prefix and branch equal, keep this branch, recurse down

					// Remove the common prefix from the key we're inserting
					newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
					newEncodedPathForRecursion := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

					// Recurse down the left branch with the remaining path
					_, err = t.insert(key, data, currentRoot.LeftNode, newEncodedPathForRecursion)
					if err != nil {
						return nil, err
					}

					// Recalculate hash after recursion
					currentRoot.CalculateBranchHash(t.HashAlgo)
				} else {
					// branches differ, create new branch and move existing branch to correct side of new branch, then recurse
					newBranchNode := &Node{
						Data:      nil,
						LeftNode:  nil,
						RightNode: nil,
						Path:      branchNodePath,
						IsLeaf:    false,
						Key:       nil,
					}

					// Calculate remaining paths after common prefix
					newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
					newNodePath := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

					oldBranchKeyCut, oldBranchKeyBitLen := removeFirstNBitsWithLen(leftKey, leftMeaningfulBits, commonPrefixLen)
					oldBranchNewPath := EncodeKeyBitsAsPath(oldBranchKeyCut, oldBranchKeyBitLen)

					// Update the existing branch's path
					currentRoot.LeftNode.Path = oldBranchNewPath
					currentRoot.LeftNode.CalculateBranchHash(t.HashAlgo)

					// Determine which side each goes on based on the next bit after common prefix
					if utils.GetBit(leftKey, commonPrefixLen) {
						// old branch goes right, new node goes left
						newBranchNode.RightNode = currentRoot.LeftNode
						newBranchNode.LeftNode = &Node{
							Data:      nodeDataCbor,
							LeftNode:  nil,
							RightNode: nil,
							Path:      newNodePath,
							IsLeaf:    true,
							Key:       nodeKeyHash,
							Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
						}
					} else {
						// old branch stays left, new node goes right
						newBranchNode.LeftNode = currentRoot.LeftNode
						newBranchNode.RightNode = &Node{
							Data:      nodeDataCbor,
							LeftNode:  nil,
							RightNode: nil,
							Path:      newNodePath,
							IsLeaf:    true,
							Key:       nodeKeyHash,
							Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
						}
					}

					newBranchNode.CalculateBranchHash(t.HashAlgo)
					currentRoot.LeftNode = newBranchNode
					currentRoot.CalculateBranchHash(t.HashAlgo)
				}
			}
		}
	} else {
		// check right of current root
		if currentRoot.RightNode == nil {
			// no node at all
			var path []byte
			if newEncodedPath == nil {
				// not in recursion
				path = EncodeKeyBitsAsPath(nodeKeyHash, len(nodeKeyHash)*8)
			} else {
				// in recursion
				path = newEncodedPath
			}
			currentRoot.RightNode = &Node{
				Data:      nodeDataCbor,
				LeftNode:  nil,
				RightNode: nil,
				Path:      path,
				Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, path, nodeDataCbor),
				IsLeaf:    true,
				Key:       nodeKeyHash,
			}
			currentRoot.CalculateBranchHash(t.HashAlgo)
		} else {
			// have node at right side
			if currentRoot.RightNode.IsLeaf {
				// right is leaf, create branch
				// find common prefix of inserted and current leaf on the right
				rightKey, rightMeaningfulBits := CalculateKeyFromPath(currentRoot.RightNode.Path)
				commonPrefix, commonPrefixLen, _ := utils.FindCommonBitPrefixWithLen(rightKey, rightMeaningfulBits, chosenKey, chosenKeyBitLen)
				branchNodePath := EncodeKeyBitsAsPath(commonPrefix, commonPrefixLen)
				branchNode := &Node{
					Data:      nil,
					LeftNode:  nil,
					RightNode: nil,
					Path:      branchNodePath,
					IsLeaf:    false,
					Key:       nil,
				}

				newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
				newNodePath := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

				oldNodeKeyCut, oldNodeKeyBitLen := removeFirstNBitsWithLen(rightKey, rightMeaningfulBits, commonPrefixLen)
				oldNodeNewPath := EncodeKeyBitsAsPath(oldNodeKeyCut, oldNodeKeyBitLen)
				currentRoot.RightNode.Path = oldNodeNewPath
				currentRoot.RightNode.Hash = utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, currentRoot.RightNode.Path, currentRoot.RightNode.Data)

				if utils.GetBit(rightKey, commonPrefixLen) {
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
				branchNode.CalculateBranchHash(t.HashAlgo)
				currentRoot.RightNode = branchNode
				currentRoot.CalculateBranchHash(t.HashAlgo)
			} else {
				// right is branch
				// 1. find common prefix of inserting node and existing branch node
				rightKey, rightMeaningfulBits := CalculateKeyFromPath(currentRoot.RightNode.Path)
				commonPrefix, commonPrefixLen, _ := utils.FindCommonBitPrefixWithLen(rightKey, rightMeaningfulBits, chosenKey, chosenKeyBitLen)
				// 2. check if path from prefix is equal to branch's path
				branchNodePath := EncodeKeyBitsAsPath(commonPrefix, commonPrefixLen)
				if bytes.Equal(branchNodePath, currentRoot.RightNode.Path) {
					// prefix and branch equal, keep this branch, recurse down

					// Remove the common prefix from the key we're inserting
					newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
					newEncodedPathForRecursion := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

					// Recurse down the right branch with the remaining path
					_, err = t.insert(key, data, currentRoot.RightNode, newEncodedPathForRecursion)
					if err != nil {
						return nil, err
					}

					// Recalculate hash after recursion
					currentRoot.CalculateBranchHash(t.HashAlgo)
				} else {
					// branches differ, create new branch and move existing branch to correct side of new branch, then recurse
					newBranchNode := &Node{
						Data:      nil,
						LeftNode:  nil,
						RightNode: nil,
						Path:      branchNodePath,
						IsLeaf:    false,
						Key:       nil,
					}

					// Calculate remaining paths after common prefix
					newNodeKeyCut, newNodeKeyBitLen := removeFirstNBitsWithLen(chosenKey, chosenKeyBitLen, commonPrefixLen)
					newNodePath := EncodeKeyBitsAsPath(newNodeKeyCut, newNodeKeyBitLen)

					oldBranchKeyCut, oldBranchKeyBitLen := removeFirstNBitsWithLen(rightKey, rightMeaningfulBits, commonPrefixLen)
					oldBranchNewPath := EncodeKeyBitsAsPath(oldBranchKeyCut, oldBranchKeyBitLen)

					// Update the existing branch's path
					currentRoot.RightNode.Path = oldBranchNewPath
					currentRoot.RightNode.CalculateBranchHash(t.HashAlgo)

					// Determine which side each goes on based on the next bit after common prefix
					if utils.GetBit(rightKey, commonPrefixLen) {
						// old branch stays right, new node goes left
						newBranchNode.RightNode = currentRoot.RightNode
						newBranchNode.LeftNode = &Node{
							Data:      nodeDataCbor,
							LeftNode:  nil,
							RightNode: nil,
							Path:      newNodePath,
							IsLeaf:    true,
							Key:       nodeKeyHash,
							Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
						}
					} else {
						// old branch goes left, new node goes right
						newBranchNode.LeftNode = currentRoot.RightNode
						newBranchNode.RightNode = &Node{
							Data:      nodeDataCbor,
							LeftNode:  nil,
							RightNode: nil,
							Path:      newNodePath,
							IsLeaf:    true,
							Key:       nodeKeyHash,
							Hash:      utils.ConcatDataAndGenerateCombinedHash(t.HashAlgo, newNodePath, nodeDataCbor),
						}
					}

					newBranchNode.CalculateBranchHash(t.HashAlgo)
					currentRoot.RightNode = newBranchNode
					currentRoot.CalculateBranchHash(t.HashAlgo)
				}
			}
		}
	}

	return currentRoot, nil
}
