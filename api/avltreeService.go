package api

import (
	"encoding/hex"
	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TreeResponse struct {
	Root *NodeResponse `json:"root,omitempty"`
}

type NodeResponse struct {
	Key         string        `json:"key"` // Hex-encoded key
	KeyDecoded  interface{}   `json:"keyDecoded"`
	Data        interface{}   `json:"data"` // Raw Data
	DataDecoded interface{}   `json:"dataDecoded"`
	Height      int           `json:"height"`
	NodeHash    string        `json:"nodeHash"`    // Hex-encoded hash
	SubtreeHash string        `json:"subtreeHash"` // Hex-encoded hash
	LeftChild   *NodeResponse `json:"leftChild,omitempty"`
	RightChild  *NodeResponse `json:"rightChild,omitempty"`
}

type InsertRequest struct {
	Key  string      `json:"key" binding:"required"`  // Hex-encoded key
	Data interface{} `json:"data" binding:"required"` // Data to store
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

type AVLHashTreeService struct {
	AVLHashTree *avlhashtree.AVLHashTree
}

func (s *AVLHashTreeService) convertNodeToResponse(node *avlhashtree.Node) *NodeResponse {
	if node == nil {
		return nil
	}

	// Decode the data for response, try different types
	var decodedData interface{}
	if intData, err := utils.DecodeCBOR[int](node.Data); err == nil {
		decodedData = intData
	} else if strData, err := utils.DecodeCBOR[string](node.Data); err == nil {
		decodedData = strData
	} else if floatData, err := utils.DecodeCBOR[float64](node.Data); err == nil {
		decodedData = floatData
	} else if byteData, err := utils.DecodeCBOR[[]byte](node.Data); err == nil {
		decodedData = hex.EncodeToString(byteData)
	} else {
		decodedData = hex.EncodeToString(node.Data)
	}

	// Decode the key for response, try different types
	var decodedKey interface{}
	if intKey, err := utils.DecodeCBOR[int](node.Key); err == nil {
		decodedKey = intKey
	} else if strKey, err := utils.DecodeCBOR[string](node.Key); err == nil {
		decodedKey = strKey
	} else if floatKey, err := utils.DecodeCBOR[float64](node.Key); err == nil {
		decodedKey = floatKey
	} else if byteKey, err := utils.DecodeCBOR[[]byte](node.Key); err == nil {
		decodedKey = hex.EncodeToString(byteKey)
	} else {
		decodedKey = hex.EncodeToString(node.Key)
	}

	return &NodeResponse{
		Key:         hex.EncodeToString(node.Key),
		KeyDecoded:  decodedKey,
		Data:        hex.EncodeToString(node.Data),
		DataDecoded: decodedData,
		Height:      node.Height,
		NodeHash:    hex.EncodeToString(node.NodeHash),
		SubtreeHash: hex.EncodeToString(node.SubtreeHash),
		LeftChild:   s.convertNodeToResponse(node.LeftChild),
		RightChild:  s.convertNodeToResponse(node.RightChild),
	}
}

func NewAVLHashTreeService(tree *avlhashtree.AVLHashTree) *AVLHashTreeService {
	return &AVLHashTreeService{AVLHashTree: tree}
}

func SetupRoutesWithService(r *gin.Engine, avlHashTreeService *AVLHashTreeService) {
	r.GET("/avltree", avlHashTreeService.GetAVLHashTree)                // Get whole tree
	r.GET("/avltree/:key", avlHashTreeService.GetAVLHashTreeNode)       // Get single node value
	r.POST("/avltree", avlHashTreeService.InsertAVLHashTreeNode)        // Insert key and data
	r.DELETE("/avltree/:key", avlHashTreeService.DeleteAVLHashTreeNode) // Delete key from tree
}

// AVL Hash Tree CRUD ops

func (s *AVLHashTreeService) GetAVLHashTree(c *gin.Context) {
	response := TreeResponse{
		Root: s.convertNodeToResponse(s.AVLHashTree.Root),
	}

	c.JSON(http.StatusOK, response)
}

func (s *AVLHashTreeService) GetAVLHashTreeNode(c *gin.Context) {
	keyParam := c.Param("key")
	if keyParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_key",
			Message: "Key parameter is required",
		})
		return
	}

	// Decode hex key to bytes
	keyBytes, err := hex.DecodeString(keyParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_key_format",
			Message: "Key must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	// Encode key as CBOR for search
	keyCBOR, err := utils.EncodeCBOR(keyBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "encoding_error",
			Message: "Failed to encode key",
			Stack:   err.Error(),
		})
		return
	}

	// Search for the node
	data, err := s.AVLHashTree.Search(keyCBOR)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "node_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":  keyParam,
		"data": data,
	})
}

func (s *AVLHashTreeService) InsertAVLHashTreeNode(c *gin.Context) {
	var req InsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Decode hex key to bytes
	keyBytes, err := hex.DecodeString(req.Key)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_key_format",
			Message: "Key must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	// Convert key bytes to Hash type
	keyHash := utils.Hash(keyBytes)

	// Insert into tree
	err = s.AVLHashTree.Insert(keyHash, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "insert_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Node inserted successfully",
		"key":     req.Key,
		"data":    req.Data,
	})
}

func (s *AVLHashTreeService) DeleteAVLHashTreeNode(c *gin.Context) {
	keyParam := c.Param("key")
	if keyParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_key",
			Message: "Key parameter is required",
		})
		return
	}

	// Decode hex key to bytes
	keyBytes, err := hex.DecodeString(keyParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_key_format",
			Message: "Key must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	// Convert key bytes to Hash type
	keyHash := utils.Hash(keyBytes)

	// Delete from tree
	err = s.AVLHashTree.Delete(keyHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Node deleted successfully",
		"key":     keyParam,
	})
}
