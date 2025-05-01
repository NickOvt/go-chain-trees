package avlhashtree

import (
	"bytes"
	"crypto/sha3"
	"errors"
	"fmt"
	"strings"
)

// Node represents a node in the AVL tree
type Node struct {
	Key        []byte // Hash value used as key
	Data       []byte // Original data or reference
	Height     int    // Height used for balancing
	LeftChild  *Node
	RightChild *Node
}

// Main AVL Tree struct
type AVLHashTree struct {
	Root *Node
}

func GenerateHash(data []byte) []byte {
	hash := sha3.Sum384(data)
	return hash[:]
}

// NewAVLHashTree creates a new, empty AVL hash tree
func NewAVLHashTree() *AVLHashTree {
	return &AVLHashTree{Root: nil}
}

func height(node *Node) int {
	if node == nil {
		return 0
	}
	return node.Height
}

func getBalanceFactor(node *Node) int {
	if node == nil {
		return 0
	}
	return height(node.LeftChild) - height(node.RightChild)
}

func getMinNode(node *Node) *Node {
	if node == nil || node.LeftChild == nil {
		return nil
	}
	return getMinNode(node.LeftChild)
}

func (t *AVLHashTree) Search(key []byte) ([]byte, error) {
	node := t.search(t.Root, key)
	if node == nil {
		return nil, errors.New("Node with given hashkey not found")
	}
	return node.Data, nil
}

func (t *AVLHashTree) search(node *Node, key []byte) *Node {
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

	return A
}

func (t *AVLHashTree) Insert(key []byte) error {
	t.Root = t.insert(t.Root, key)
	return nil
}

func (t *AVLHashTree) insert(root *Node, key []byte) *Node {
	if root == nil {
		return &Node{
			Key:        key,
			Data:       nil,
			Height:     1,
			LeftChild:  nil,
			RightChild: nil,
		}
	}

	cmp := bytes.Compare(key, root.Key)

	if cmp < 0 {
		root.LeftChild = t.insert(root.LeftChild, key)
	} else if cmp > 0 {
		root.RightChild = t.insert(root.RightChild, key)
	} else {
		// duplicate
		root.RightChild = t.insert(root.RightChild, key)
	}

	root.Height = 1 + max(height(root.LeftChild), height(root.RightChild))

	balanceFactor := getBalanceFactor(root)

	if balanceFactor > 1 && bytes.Compare(key, root.LeftChild.Key) < 0 {
		return t.rotateRight(root)
	}

	if balanceFactor < -1 && bytes.Compare(key, root.RightChild.Key) > 0 {
		return t.rotateLeft(root)
	}

	if balanceFactor > 1 && bytes.Compare(key, root.LeftChild.Key) > 0 {
		root.LeftChild = t.rotateLeft(root.LeftChild)
		return t.rotateRight(root)
	}

	if balanceFactor < -1 && bytes.Compare(key, root.RightChild.Key) < 0 {
		root.RightChild = t.rotateRight(root.RightChild)
		return t.rotateLeft(root)
	}

	return root
}

func (t *AVLHashTree) Delete(key []byte) error {
	t.Root = t.delete(t.Root, key)
	return nil
}

func (t *AVLHashTree) delete(root *Node, key []byte) *Node {
	if root == nil {
		return nil
	}

	cmp := bytes.Compare(key, root.Key)

	if cmp < 0 {
		root.LeftChild = t.delete(root.LeftChild, key)
	} else if cmp > 0 {
		root.RightChild = t.delete(root.RightChild, key)
	} else {
		if root.LeftChild == nil {
			temp := root.RightChild
			root = nil
			return temp
		} else if root.RightChild == nil {
			temp := root.LeftChild
			root = nil
			return temp
		}

		// Find inorder successor
		temp := getMinNode(root.RightChild)
		root.Key = temp.Key
		root.RightChild = t.delete(root.RightChild, temp.Key)
	}

	root.Height = 1 + max(height(root.LeftChild), height(root.RightChild))

	balanceFactor := getBalanceFactor(root)

	if balanceFactor > 1 && getBalanceFactor(root.LeftChild) >= 0 {
		return t.rotateRight(root)
	}

	if balanceFactor < -1 && getBalanceFactor(root.RightChild) <= 0 {
		return t.rotateLeft(root)
	}

	if balanceFactor > 1 && getBalanceFactor(root.LeftChild) < 0 {
		root.LeftChild = t.rotateLeft(root.LeftChild)
		return t.rotateRight(root)
	}

	if balanceFactor < -1 && getBalanceFactor(root.RightChild) > 0 {
		root.RightChild = t.rotateRight(root.RightChild)
		return t.rotateLeft(root)
	}

	return root
}

func (t *AVLHashTree) PrintTree() {
	if t.Root != nil {
		// Create a queue for level-order traversal
		queue := []*Node{t.Root}
		levelOrder := strings.Builder{}
		levelOrderWithDetails := strings.Builder{}

		// Process each node in the queue
		for len(queue) > 0 {
			// Dequeue the front node
			node := queue[0]
			queue = queue[1:]

			// Add the key to the level-order string
			levelOrder.WriteString(fmt.Sprintf("%x ", node.Key[0:4]))

			// Add the key with height and balance factor to the detailed string
			levelOrderWithDetails.WriteString(fmt.Sprintf("%x(h=%d, bf=%d) ",
				node.Key[0:4], node.Height, getBalanceFactor(node)))

			// Add children to the queue
			if node.LeftChild != nil {
				queue = append(queue, node.LeftChild)
			}
			if node.RightChild != nil {
				queue = append(queue, node.RightChild)
			}
		}

		// Print the results
		fmt.Println("\nLevel-order traversal:")
		fmt.Println(levelOrder.String())
		fmt.Println("\nLevel-order traversal with height and balance factor:")
		fmt.Println(levelOrderWithDetails.String())
	} else {
		fmt.Println("\nAVL tree is empty!")
	}
}
