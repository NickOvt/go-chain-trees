package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
)

func main() {
	log.Println("Starting application...")

	// Create a new AVL tree
	avl := avlhashtree.NewAVLHashTree()

	// Insert keys
	keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	for _, key := range keys {
		avl.Insert(utils.GenerateHash([]byte(strconv.Itoa(key))), key)
		// avl.PrintTree()
	}

	// Print the final tree
	avl.PrintTree()

	search, _ := utils.EncodeCBOR(utils.GenerateHash([]byte(strconv.Itoa(4))))

	fmt.Println(fmt.Sprintf("%x|%x", search, utils.GenerateHash([]byte(strconv.Itoa(4)))))

	aaa, _ := utils.DecodeCBOR[utils.Hash](search)

	fmt.Println(fmt.Sprintf("%x", aaa))

	fmt.Println(avl.Search(search))
}
