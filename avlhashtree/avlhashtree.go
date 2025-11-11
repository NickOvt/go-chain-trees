package avlhashtree

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/NickOvt/go-chain-trees/utils"
)

// Node represents a node in the AVL tree.
// each node stores CBOR-encoded data and maintains cryptographic hashes for integrity verification.
// The node contains both its own hash and a subtree hash that includes all descendant nodes.
type Node struct {
	Key         utils.CBORData // Hash value used as key = ID, CBOR
	Data        utils.CBORData // Original data, CBOR
	Height      int            // Height used for balancing
	LeftChild   *Node
	RightChild  *Node
	NodeHash    utils.Hash // Hash of Key + Data (Current node's hash)
	SubtreeHash utils.Hash // NodeHash + SubtreeHash of Left Child + SubtreeHash of Right Child
}

// Returns the NodeHash of the current node.
// Safely handles nil nodes by returning nil.
//
// Returns:
//   - utils.Hash: The node's hash, or nil if the node is nil
func (node *Node) getNodeHash() utils.Hash {
	if node == nil {
		return nil
	}

	return node.NodeHash
}

func (node *Node) getNodeSubtreeHash() utils.Hash {
	if node == nil {
		return nil
	}

	return node.SubtreeHash
}

// Computes and updates the SubtreeHash for the current node.
// The subtree hash is calculated by combining the current node's hash with the
// hashes of its left and right children.
//
// Returns:
//   - utils.Hash: The calculated subtree hash
//   - error: An error if CBOR encoding fails, nil otherwise
func (node *Node) calculateSubtreeHash() (utils.Hash, error) {
	encodedCBORList, err := utils.EncodeCBORList(node.getNodeHash(), node.LeftChild.getNodeSubtreeHash(), node.RightChild.getNodeSubtreeHash())

	if err != nil {
		return nil, err
	}

	node.SubtreeHash = utils.ConcatDataAndGenerateHash(encodedCBORList...)

	return node.SubtreeHash, nil
}

// AVLHashTree Main AVL Tree struct
type AVLHashTree struct {
	Root *Node
}

// NewAVLHashTree creates a new, empty AVL hash tree
//
// Returns:
//   - *AVLHashTree: Empty AVLHashTree struct
func NewAVLHashTree() *AVLHashTree {
	return &AVLHashTree{Root: nil}
}

// Returns the height of the given node in the tree.
//
// Parameters:
//   - node: The *Node pointer whose height to retrieve
//
// Returns:
//   - int: The height of the node, or 0 if the node is nil
func height(node *Node) int {
	if node == nil {
		return 0
	}
	return node.Height
}

// Calculates the balance factor of the given node.
//
// Parameters:
//   - node: The *Node pointer whose balance factor to calculate
//
// Returns:
//   - int: The balance factor of the node, or 0 if the node is nil
func getBalanceFactor(node *Node) int {
	if node == nil {
		return 0
	}
	return height(node.LeftChild) - height(node.RightChild)
}

// Finds and returns the node with the minimum key in the subtree
// rooted at the given node.
//
// Parameters:
//   - node: The root *Node pointer of the subtree to search
//
// Returns:
//   - *Node: The node with the minimum key, or nil if the input node is nil
func getMinNode(node *Node) *Node {
	if node == nil {
		return nil
	}

	currNode := node
	for currNode.LeftChild != nil {
		currNode = currNode.LeftChild
	}
	return currNode
}

// Search for node in AVL Hash Tree
//
// Params:
// Key: utils.CBORData CBOR Data
//
// Returns:
// Node data as the provided decode datatype or an error
func (t *AVLHashTree) Search(key utils.CBORData) (any, error) {
	node := t.search(t.Root, key)
	if node == nil {
		return nil, errors.New("node with given hashkey not found")
	}

	dataDecoded, err := utils.DecodeCBOR[any](node.Data)

	if err != nil {
		return nil, err
	}

	return dataDecoded, nil
}

func (t *AVLHashTree) search(node *Node, key utils.CBORData) *Node {
	for node != nil {
		cmp := bytes.Compare(key, node.Key)

		if cmp == 0 {
			return node
		} else if cmp < 0 {
			node = node.LeftChild
		} else {
			node = node.RightChild
		}
	}
	return node
}

func (t *AVLHashTree) rotateLeft(node *Node) *Node {
	B := node.RightChild
	Y := B.LeftChild

	// Perform rotation
	B.LeftChild = node
	node.RightChild = Y

	// Update heights
	node.Height = 1 + max(height(node.LeftChild), height(node.RightChild))
	B.Height = 1 + max(height(B.LeftChild), height(B.RightChild))

	node.calculateSubtreeHash()
	B.calculateSubtreeHash()

	return B
}

func (t *AVLHashTree) rotateRight(node *Node) *Node {
	A := node.LeftChild
	Y := A.RightChild

	// Perform rotation
	A.RightChild = node
	node.LeftChild = Y

	// Update heights
	node.Height = 1 + max(height(node.LeftChild), height(node.RightChild))
	A.Height = 1 + max(height(A.LeftChild), height(A.RightChild))

	node.calculateSubtreeHash() // Old root first
	A.calculateSubtreeHash()

	return A
}

// InsertCBOR Insert CBOR data directly into the AVL Hash Tree
//
// Params:
//   - keyCBOR: utils.CBORData key encoded as CBOR
//   - dataCBOR: utils.CBORData data encoded as CBOR
//
// Returns:
//   - nil
func (t *AVLHashTree) InsertCBOR(keyCBOR utils.CBORData, dataCBOR utils.CBORData) error {
	// Encode the hash bytes to CBOR
	t.Root = t.insert(t.Root, keyCBOR, dataCBOR)
	return nil
}

// Insert data into the AVL Hash Tree.
// Hash and Data are converted to CBOR before inserting
//
// Params:
//   - key: utils.Hash of key
//   - data: any data
//
// Returns:
//   - nil
func (t *AVLHashTree) Insert(key utils.Hash, data any) error {
	// Encode the hash bytes to CBOR
	encodedKey, err := utils.EncodeCBOR(key)
	if err != nil {
		return err
	}

	// Encode data to CBOR only if it's not nil
	var dataCBOR utils.CBORData
	if data != nil {
		var err error
		dataCBOR, err = utils.EncodeCBOR(data)
		if err != nil {
			return err
		}
	}

	t.Root = t.insert(t.Root, encodedKey, dataCBOR)
	return nil
}

func (t *AVLHashTree) insert(root *Node, key utils.CBORData, data utils.CBORData) *Node {
	newRoot, _ := t.insertRecursive(root, key, data)
	return newRoot
}

func (t *AVLHashTree) insertRecursive(root *Node, key utils.CBORData, data utils.CBORData) (*Node, bool) {
	if root == nil {
		node := &Node{
			Key:        key,
			Data:       data,
			Height:     1,
			LeftChild:  nil,
			RightChild: nil,
			NodeHash:   utils.ConcatDataAndGenerateHash(key, data),
		}
		node.calculateSubtreeHash()
		return node, true

	}

	var changed bool
	cmp := bytes.Compare(key, root.Key)

	if cmp < 0 {
		root.LeftChild, changed = t.insertRecursive(root.LeftChild, key, data)
	} else if cmp > 0 {
		root.RightChild, changed = t.insertRecursive(root.RightChild, key, data)
	} else {
		// duplicate
		possibleNewNodeHash := utils.ConcatDataAndGenerateHash(key, data)
		if !bytes.Equal(root.getNodeHash(), possibleNewNodeHash) {
			root.Data = data
			root.NodeHash = possibleNewNodeHash
			changed = true
		}
	}

	if !changed {
		return root, false
	}

	root.Height = 1 + max(height(root.LeftChild), height(root.RightChild))

	balanceFactor := getBalanceFactor(root)

	if balanceFactor > 1 && bytes.Compare(key, root.LeftChild.Key) < 0 {
		root = t.rotateRight(root)
	}

	if balanceFactor < -1 && bytes.Compare(key, root.RightChild.Key) > 0 {
		root = t.rotateLeft(root)
	}

	if balanceFactor > 1 && bytes.Compare(key, root.LeftChild.Key) > 0 {
		root.LeftChild = t.rotateLeft(root.LeftChild)
		root = t.rotateRight(root)
	}

	if balanceFactor < -1 && bytes.Compare(key, root.RightChild.Key) < 0 {
		root.RightChild = t.rotateRight(root.RightChild)
		root = t.rotateLeft(root)
	}

	// Recalculate Subtree Hash for root here, after insert and rotations
	_, err := root.calculateSubtreeHash()
	if err != nil {
		// handle error
		fmt.Println("Calculate Subtree Hash Failed")
	}

	return root, true
}

// DeleteCBOR Deletes a node from the AVL Hash Tree based on the provided key
//
// Params:
//   - key: utils.CBORData key of the node to delete encoded as CBOR
//
// Returns:
//   - nil
func (t *AVLHashTree) DeleteCBOR(key utils.CBORData) error {
	t.Root = t.delete(t.Root, key)
	return nil
}

// Delete Deletes a node from the AVL Hash Tree based on the provided key
//
// Params:
//   - key: raw utils.Hash of key of the node to delete
//
// Returns:
//   - nil
func (t *AVLHashTree) Delete(key utils.Hash) error {
	encodedKey, err := utils.EncodeCBOR(key)
	if err != nil {
		return err
	}

	t.Root = t.delete(t.Root, encodedKey)
	return nil
}

func (t *AVLHashTree) delete(root *Node, key utils.CBORData) *Node {
	newRoot, _ := t.deleteRecursive(root, key)
	return newRoot
}

func (t *AVLHashTree) deleteRecursive(root *Node, key utils.CBORData) (*Node, bool) {
	if root == nil {
		return nil, false
	}

	var changed bool
	cmp := bytes.Compare(key, root.Key)

	if cmp < 0 {
		root.LeftChild, changed = t.deleteRecursive(root.LeftChild, key)
	} else if cmp > 0 {
		root.RightChild, changed = t.deleteRecursive(root.RightChild, key)
	} else {
		if root.LeftChild == nil {
			temp := root.RightChild
			root = nil
			return temp, true
		} else if root.RightChild == nil {
			temp := root.LeftChild
			root = nil
			return temp, true
		}

		// Find inorder successor
		temp := getMinNode(root.RightChild)
		root.Key = temp.Key
		root.Data = temp.Data
		root.NodeHash = temp.NodeHash
		root.RightChild, changed = t.deleteRecursive(root.RightChild, temp.Key)
	}

	if !changed {
		return root, false
	}

	root.Height = 1 + max(height(root.LeftChild), height(root.RightChild))

	balanceFactor := getBalanceFactor(root)

	if balanceFactor > 1 && getBalanceFactor(root.LeftChild) >= 0 {
		root = t.rotateRight(root)
	}

	if balanceFactor < -1 && getBalanceFactor(root.RightChild) <= 0 {
		root = t.rotateLeft(root)
	}

	if balanceFactor > 1 && getBalanceFactor(root.LeftChild) < 0 {
		root.LeftChild = t.rotateLeft(root.LeftChild)
		root = t.rotateRight(root)
	}

	if balanceFactor < -1 && getBalanceFactor(root.RightChild) > 0 {
		root.RightChild = t.rotateRight(root.RightChild)
		root = t.rotateLeft(root)
	}

	_, err := root.calculateSubtreeHash()
	if err != nil {
		// handle error
		fmt.Println("Calculate Subtree Hash Failed")
	}

	return root, true
}

// PrintTree prints a visual representation of the AVL tree structure to stdout.
// It performs a level-order (breadth-first) traversal and displays each level
// of the tree on a separate line. For each node, it shows:
// - The first 4 bytes of the node's key in hexadecimal format
// - The node's height (h=)
// - The node's balance factor (bf=)
//
// If the tree is empty, it prints a message indicating so.
//
// Example output:
//
//	Level 0: a1b2c3d4 (h=2, bf=0)
//	Level 1: e5f6a7b8 (h=1, bf=-1)  c9d10e11 (h=1, bf=0)
func (t *AVLHashTree) PrintTree() {
	if t.Root == nil {
		fmt.Println("\nAVL tree is empty!")
		return
	}

	// Create a queue for level-order traversal
	queue := []*Node{t.Root}
	level := 0

	fmt.Println("\nAVL Tree Structure:")
	for len(queue) > 0 {
		// Number of nodes at the current level
		levelSize := len(queue)
		fmt.Printf("Level %d: ", level)

		// Process all nodes at the current level
		for i := 0; i < levelSize; i++ {
			// Dequeue the front node
			node := queue[0]
			queue = queue[1:]

			// Print the node's key (first 4 bytes for brevity)
			fmt.Printf("%x (h=%d, bf=%d)  ", node.Key[0:4], node.Height, getBalanceFactor(node))

			// Enqueue children
			if node.LeftChild != nil {
				queue = append(queue, node.LeftChild)
			}
			if node.RightChild != nil {
				queue = append(queue, node.RightChild)
			}
		}

		fmt.Println() // Move to the next line for the next level
		level++
	}
}

// ValidateTree performs a comprehensive validation of the AVL tree structure
// and cryptographic integrity. It verifies that all nodes maintain proper
// hash relationships and that the hashes are intact.
// The validation checks include:
//   - NodeHash integrity (Key + Data hashing)
//   - SubtreeHash integrity (node hash + children hashes)
//   - Recursive validation of all nodes in the tree
//
// An empty tree (nil root) is considered valid.
//
// Returns:
//   - error: nil if the tree is valid, otherwise an error describing the
//     first validation failure encountered with details about which
//     node failed and what type of hash mismatch occurred
func (t *AVLHashTree) ValidateTree() error {
	if t.Root == nil {
		return nil // Empty tree is valid
	}

	err := t.validateNode(t.Root)

	if err != nil {
		return err
	}

	fmt.Println("Tree validation succeeded")
	return nil
}

// validateNode recursively validates a node and all its children
func (t *AVLHashTree) validateNode(node *Node) error {
	if node == nil {
		return nil
	}

	// 1. Validate NodeHash (Key + Data)
	expectedNodeHash := utils.ConcatDataAndGenerateHash(node.Key, node.Data)
	if !bytes.Equal(node.NodeHash, expectedNodeHash) {
		return fmt.Errorf("invalid NodeHash for key %x: expected %x, got %x",
			node.Key[:min(len(node.Key), 8)], expectedNodeHash, node.NodeHash)
	}

	// 2. Recursively validate children first
	if err := t.validateNode(node.LeftChild); err != nil {
		return err
	}
	if err := t.validateNode(node.RightChild); err != nil {
		return err
	}

	// 3. Validate SubtreeHash (NodeHash + Subtree Hash of Left Child + Subtree Hash of Right Child)
	leftHash := node.LeftChild.getNodeSubtreeHash()
	rightHash := node.RightChild.getNodeSubtreeHash()

	// The SubtreeHash should include the current node's hash plus child hashes
	// First encode all three hashes to CBOR
	encodedCBORList, err := utils.EncodeCBORList(node.NodeHash, leftHash, rightHash)
	if err != nil {
		return fmt.Errorf("failed to encode node and child hashes for key %x: %v",
			node.Key[:min(len(node.Key), 8)], err)
	}

	expectedSubtreeHash := utils.ConcatDataAndGenerateHash(encodedCBORList...)
	if !bytes.Equal(node.SubtreeHash, expectedSubtreeHash) {
		return fmt.Errorf("invalid SubtreeHash for key %x: expected %x, got %x",
			node.Key[:min(len(node.Key), 8)], expectedSubtreeHash, node.SubtreeHash)
	}

	return nil
}

type CryptographicProof struct {
	RootHash  utils.Hash     // Complete hashChain
	TargetKey utils.CBORData // Key for which to generate inclusion/exclusion proof
	Found     bool           // Inclusion proved
	Path      []*Node        // Hash Chain Path
	Direction string         // If inclusion not proved specify what direction node would be required
	ChainSize int            // Size hash chain path
}

// TODO: Return Proof as CBOR bytes

func (t *AVLHashTree) GenerateInclusionExclusionProof(key utils.CBORData) (*CryptographicProof, error) {
	if t.Root == nil {
		return &CryptographicProof{
			RootHash:  nil,
			TargetKey: key,
			Found:     false,
			Path:      []*Node{},
			ChainSize: 0,
		}, nil
	}

	proof := &CryptographicProof{
		RootHash:  t.Root.getNodeSubtreeHash(),
		TargetKey: key,
		Found:     false,
		Path:      []*Node{},
		ChainSize: 0,
	}

	t.generateInclusionExclusionProofRecursive(t.Root, key, "root", proof)

	return proof, nil
}

func (t *AVLHashTree) generateInclusionExclusionProofRecursive(node *Node, targetKey utils.CBORData, direction string, proof *CryptographicProof) bool {
	if node == nil {
		proof.Direction = direction // Node not found so add child that should've been
		return false
	}

	proof.ChainSize++

	cmp := bytes.Compare(targetKey, node.Key)

	if cmp == 0 {
		proof.Found = true
		proof.Path = append(proof.Path, node)
		return true
	}

	proof.Path = append(proof.Path, node)

	if cmp < 0 {
		// Go left
		return t.generateInclusionExclusionProofRecursive(node.LeftChild, targetKey, "left", proof)
	} else {
		// Go right
		return t.generateInclusionExclusionProofRecursive(node.RightChild, targetKey, "right", proof)
	}
}

type PublicCryptographicProof struct {
	RootHash  utils.Hash                      `cbor:"1,keyasint"`
	TargetKey utils.CBORData                  `cbor:"2,keyasint"`
	Found     bool                            `cbor:"3,keyasint"`
	Path      []*PublicCryptographicProofNode `cbor:"4,keyasint"`
	Direction string                          `cbor:"5,keyasint, omitempty"`
}

type PublicCryptographicProofNode struct {
	Key                   utils.CBORData `cbor:"1,keyasint"`
	Data                  utils.CBORData `cbor:"2,keyasint"`
	NodeHash              utils.Hash     `cbor:"3,keyasint"`
	LeftChildSubtreeHash  utils.Hash     `cbor:"4,keyasint, omitempty"`
	RightChildSubtreeHash utils.Hash     `cbor:"5,keyasint, omitempty"`
}

func (proof *CryptographicProof) ToPublicProof() *PublicCryptographicProof {
	nodes := make([]*PublicCryptographicProofNode, proof.ChainSize)

	for i, node := range proof.Path {
		nodes[i] = &PublicCryptographicProofNode{
			Key:      node.Key,
			Data:     node.Data,
			NodeHash: node.NodeHash,
		}

		if i < len(proof.Path)-1 {
			nextNode := proof.Path[i+1]
			cmp := bytes.Compare(nextNode.Key, node.Key)

			if cmp > 0 {
				// next node is right side
				// add missing left child
				if node.LeftChild != nil && node.LeftChild.getNodeSubtreeHash() != nil {
					nodes[i].LeftChildSubtreeHash = node.LeftChild.getNodeSubtreeHash()
				}
			} else {
				// next node is left
				// add missing right child
				if node.RightChild != nil && node.RightChild.getNodeSubtreeHash() != nil {
					nodes[i].RightChildSubtreeHash = node.RightChild.getNodeSubtreeHash()
				}
			}
		} else {
			// leaf node, save all children
			if node.LeftChild != nil && node.LeftChild.getNodeSubtreeHash() != nil {
				nodes[i].LeftChildSubtreeHash = node.LeftChild.getNodeSubtreeHash()
			}

			if node.RightChild != nil && node.RightChild.getNodeSubtreeHash() != nil {
				nodes[i].RightChildSubtreeHash = node.RightChild.getNodeSubtreeHash()
			}
		}
	}

	return &PublicCryptographicProof{
		RootHash:  proof.RootHash,
		TargetKey: proof.TargetKey,
		Found:     proof.Found,
		Path:      nodes,
		Direction: proof.Direction,
	}
}

func VerifyPublicProof(proof *PublicCryptographicProof) (bool, error) {

	if len(proof.Path) == 0 || proof.RootHash == nil || !proof.Found {
		return false, nil
	}

	if !verifyPublicPathConsistency(proof) {
		return false, fmt.Errorf("invalid path")
	}

	if !proof.Found {
		if !verifyPublicExclusionConditions(proof) {
			return false, fmt.Errorf("invalid exclusion proof")
		}
	}

	return verifyPublicHashChain(proof)
}

func verifyPublicPathConsistency(proof *PublicCryptographicProof) bool {
	for i, node := range proof.Path {
		cmp := bytes.Compare(proof.TargetKey, node.Key)

		if cmp == 0 {
			// found the target
			return proof.Found
		}

		if i < len(proof.Path)-1 {
			nextNode := proof.Path[i+1]

			if cmp < 0 {
				// next node should also be < current (going left)
				if bytes.Compare(nextNode.Key, node.Key) >= 0 {
					return false
				}
			} else {
				// next node should also be > current (going right)
				if bytes.Compare(nextNode.Key, node.Key) <= 0 {
					return false
				}
			}
		}
	}

	return !proof.Found
}

func verifyPublicExclusionConditions(proof *PublicCryptographicProof) bool {
	if len(proof.Path) == 0 {
		return true
	}

	lastNode := proof.Path[len(proof.Path)-1]
	cmp := bytes.Compare(proof.TargetKey, lastNode.Key)

	if cmp == 0 {
		// proof claims it is not found but the node actually is in path
		return false
	}

	// verify last node direction
	if cmp < 0 && proof.Direction != "left" {
		return false
	}
	if cmp > 0 && proof.Direction != "right" {
		return false
	}

	return true
}

func verifyPublicHashChain(proof *PublicCryptographicProof) (bool, error) {
	calculatedHashes := make(map[int]utils.Hash)

	for i := len(proof.Path) - 1; i >= 0; i-- {
		node := proof.Path[i]

		expectedNodeHash := utils.ConcatDataAndGenerateHash(node.Key, node.Data)
		if !bytes.Equal(node.NodeHash, expectedNodeHash) {
			return false, fmt.Errorf("invalid node hash for key %x at %d: expected %x, got %x", node.Key[:8], i, expectedNodeHash, node.NodeHash)
		}

		var leftHash, rightHash utils.Hash

		if i < len(proof.Path)-1 {
			nextNode := proof.Path[i+1]
			cmp := bytes.Compare(nextNode.Key, node.Key)

			if cmp < 0 {
				// next node is left child so use its calculated hash
				leftHash = calculatedHashes[i+1]
				// right child not in path so use provided hash
				rightHash = node.RightChildSubtreeHash
			} else {
				// next node is right child so use its calculated hash
				rightHash = calculatedHashes[i+1]
				// left child not in path so use provided hash
				leftHash = node.LeftChildSubtreeHash
			}
		} else {
			// last (leaf) node in path
			leftHash = node.LeftChildSubtreeHash
			rightHash = node.RightChildSubtreeHash
		}

		encodedCBORList, err := utils.EncodeCBORList(node.NodeHash, leftHash, rightHash)
		if err != nil {
			return false, fmt.Errorf("failed to encode node hash for key %x: %v", node.Key[:8], err)
		}

		calculatedSubtreeHash := utils.ConcatDataAndGenerateHash(encodedCBORList...)
		calculatedHashes[i] = calculatedSubtreeHash
	}

	if !bytes.Equal(proof.RootHash, calculatedHashes[0]) {
		return false, fmt.Errorf("calculated root hash does not match provided root hash")
	}

	return true, nil
}

// public
// ------------
// internal

func (t *AVLHashTree) VerifyProof(proof *CryptographicProof) (bool, error) {
	if len(proof.Path) == 0 {
		return proof.RootHash == nil && !proof.Found, nil
	}

	if !bytes.Equal(proof.RootHash, t.Root.getNodeSubtreeHash()) {
		return false, fmt.Errorf("given proof document containst invalid root hash")
	}

	if !t.verifyPathConsistency(proof) {
		return false, fmt.Errorf("path is not consistent with target key")
	}

	return t.verifyHashChain(proof)
}

/* Check that the target key is present in the path of the proof
 */
func (t *AVLHashTree) verifyPathConsistency(proof *CryptographicProof) bool {
	var lastCmp int

	for _, node := range proof.Path {
		lastCmp = bytes.Compare(proof.TargetKey, node.Key)

		if lastCmp == 0 {
			// found node
			return true
		}
	}

	// node not in path
	// get last node, and check given direction
	// lastCmp here will be the result of compare of last node in the path
	if lastCmp < 0 && proof.Direction == "left" {
		return true
	}

	if lastCmp > 0 && proof.Direction == "right" {
		return true
	}

	return false
}

func (t *AVLHashTree) verifyHashChain(proof *CryptographicProof) (bool, error) {
	// In proof's path the first node must be the root node with the root hash
	// it is checked in the wrapper functions already, if the provided
	// proof document has been spoofed/tampered with the root hash of the document
	// will not equal to the root hash of the current avl tree in memory.
	for i := len(proof.Path) - 1; i >= 0; i-- {
		node := proof.Path[i]

		expectedHash := utils.ConcatDataAndGenerateHash(node.Key, node.Data)
		if !bytes.Equal(node.NodeHash, expectedHash) {
			return false, fmt.Errorf("invalid NodeHash for key %x: expected %x, got %x", node.Key, expectedHash, node.NodeHash)
		}

		encodedCBORList, err := utils.EncodeCBORList(node.getNodeHash(), node.LeftChild.getNodeSubtreeHash(), node.RightChild.getNodeSubtreeHash())

		if err != nil {
			return false, fmt.Errorf("failed to encode hashes for verification: %v", err)
		}

		expectedSubtreeHash := utils.ConcatDataAndGenerateHash(encodedCBORList...)

		if !bytes.Equal(node.SubtreeHash, expectedSubtreeHash) {
			return false, fmt.Errorf("invalid SubtreeHash for key %x: expected %x, got %x", node.Key, expectedSubtreeHash, node.SubtreeHash)
		}
	}

	return true, nil
}
