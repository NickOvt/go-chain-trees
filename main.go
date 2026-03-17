package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/NickOvt/go-chain-trees/api"
	"github.com/gin-gonic/gin"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
)

func main() {
	log.Println("Starting application...")

	// Create a new AVL tree
	avl := avlhashtree.NewAVLHashTree(utils.SHA256)

	// Insert keys
	keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	for _, key := range keys {
		err := avl.Insert([]byte(strconv.Itoa(key)), key)
		if err != nil {
			return
		}
		// avl.PrintTree()
	}

	// Print the final tree
	avl.PrintTree()

	searchKey := []byte(strconv.Itoa(4))
	fmt.Println(fmt.Sprintf("%x", avl.HashKey(searchKey)))
	fmt.Println(avl.SearchByKey(searchKey))

	if err := avl.ValidateTree(); err != nil {
		fmt.Printf("Tree validation failed: %v\n", err)
	}

	search2 := []byte(strconv.Itoa(13))
	inclusionProof, _ := avl.GenerateInclusionExclusionProofByKey(search2)

	proofCbor, _ := utils.EncodeCBOR(inclusionProof.ToPublicProof())

	fmt.Println(fmt.Sprintf("%x", proofCbor))
	//inclusionProof.Direction = "left"

	fmt.Println("proved")
	res, verifErr := avl.VerifyProof(inclusionProof)

	fmt.Println(res, verifErr)

	r := gin.Default()

	avlHashTreeService := api.NewAVLHashTreeService(avlhashtree.NewAVLHashTree(utils.SHA256))
	api.SetupRoutes(r)
	api.SetupRoutesWithService(r, avlHashTreeService) // Add Service routes

	err := r.Run(":8080")
	if err != nil {
		fmt.Println(fmt.Errorf("An error occurred when starting server: %v\n", err))
	}
}
