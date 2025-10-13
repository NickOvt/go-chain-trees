package main

import (
	"fmt"
	"github.com/NickOvt/go-chain-trees/api"
	"github.com/gin-gonic/gin"
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
		err := avl.Insert(utils.GenerateHash([]byte(strconv.Itoa(key))), key)
		if err != nil {
			return
		}
		// avl.PrintTree()
	}

	// Print the final tree
	avl.PrintTree()

	search, _ := utils.EncodeCBOR(utils.GenerateHash([]byte(strconv.Itoa(4))))

	fmt.Println(fmt.Sprintf("%x|%x", search, utils.GenerateHash([]byte(strconv.Itoa(4)))))

	aaa, _ := utils.DecodeCBOR[utils.Hash](search)

	fmt.Println(fmt.Sprintf("%x", aaa))

	fmt.Println(avl.Search(search))

	if err := avl.ValidateTree(); err != nil {
		fmt.Printf("Tree validation failed: %v\n", err)
	}

	search2, _ := utils.EncodeCBOR(utils.GenerateHash([]byte(strconv.Itoa(13))))
	inclusionProof, _ := avl.GenerateInclusionExclusionProof(search2)

	proofCbor, _ := utils.EncodeCBOR(inclusionProof.ToPublicProof())

	fmt.Println(fmt.Sprintf("%x", proofCbor))
	//inclusionProof.Direction = "left"

	fmt.Println("proved")
	res, verifErr := avl.VerifyProof(inclusionProof)

	fmt.Println(res, verifErr)

	r := gin.Default()

	avlHashTreeService := api.NewAVLHashTreeService(avlhashtree.NewAVLHashTree())
	api.SetupRoutes(r)
	api.SetupRoutesWithService(r, avlHashTreeService) // Add Service routes

	err := r.Run(":8080")
	if err != nil {
		fmt.Println(fmt.Errorf("An error occurred when starting server: %v\n", err))
	}
}
