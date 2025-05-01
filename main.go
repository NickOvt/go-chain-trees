package main

import (
	"log"
	"strconv"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
)

func main() {
	log.Println("Starting application...")

	// Create a new AVL tree
	avl := avlhashtree.NewAVLHashTree()

	// Insert keys
	keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	for _, key := range keys {
		avl.Insert(avlhashtree.GenerateHash([]byte(strconv.Itoa(key))))
		// avl.PrintTree()
	}

	// Print the final tree
	avl.PrintTree()
}
