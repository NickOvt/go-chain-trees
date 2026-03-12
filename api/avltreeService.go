package api

import (
	"bytes"
	"encoding/hex"
	"net/http"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
	"github.com/gin-gonic/gin"
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

type ProofRequest struct {
	Key string `json:"key" binding:"required"` // Hex-encoded key
}

type ProofNodeResponse struct {
	Key                   string      `json:"key"`
	KeyDecoded            interface{} `json:"keyDecoded"`
	Data                  string      `json:"data"`
	DataDecoded           interface{} `json:"dataDecoded"`
	NodeHash              string      `json:"nodeHash"`
	LeftChildSubtreeHash  string      `json:"leftChildSubtreeHash,omitempty"`
	RightChildSubtreeHash string      `json:"rightChildSubtreeHash,omitempty"`
}

type ProofResponse struct {
	RootHash         string              `json:"rootHash,omitempty"`
	TargetKey        string              `json:"targetKey"`
	TargetKeyDecoded interface{}         `json:"targetKeyDecoded"`
	Found            bool                `json:"found"`
	Path             []ProofNodeResponse `json:"path"`
	Direction        string              `json:"direction,omitempty"`
	HashAlgo         utils.HashAlgo      `json:"hashAlgo"`
	ChainSize        int                 `json:"chainSize"`
	RootNode         *NodeResponse       `json:"rootNode,omitempty"`
}

type ProofNodeRequest struct {
	Key                   string `json:"key" binding:"required"`
	Data                  string `json:"data" binding:"required"`
	NodeHash              string `json:"nodeHash" binding:"required"`
	LeftChildSubtreeHash  string `json:"leftChildSubtreeHash,omitempty"`
	RightChildSubtreeHash string `json:"rightChildSubtreeHash,omitempty"`
}

type ProofVerificationRequest struct {
	RootHash  string             `json:"rootHash" binding:"required"`
	TargetKey string             `json:"targetKey" binding:"required"`
	Found     bool               `json:"found"`
	Path      []ProofNodeRequest `json:"path" binding:"required"`
	Direction string             `json:"direction,omitempty"`
	HashAlgo  utils.HashAlgo     `json:"hashAlgo"`
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

	return &NodeResponse{
		Key:         hex.EncodeToString(node.Key),
		KeyDecoded:  decodeKeyToBestType(node.Key),
		Data:        hex.EncodeToString(node.Data),
		DataDecoded: decodeCBORToBestType(node.Data),
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
	r.GET("/avltree/validate", avlHashTreeService.ValidateAVLHashTree)  // Validate tree integrity
	r.GET("/avltree/proof/:key", avlHashTreeService.GetAVLHashTreeProof)
	r.POST("/avltree/proof/verify", avlHashTreeService.VerifyAVLHashTreeProof)
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

	node := findNodeByKey(s.AVLHashTree.Root, utils.Hash(keyBytes))
	if node == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "node_not_found",
			Message: "Node with given hashkey not found",
		})
		return
	}

	c.JSON(http.StatusOK, s.convertNodeToResponse(node))
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

	// Insert into tree
	err = s.AVLHashTree.Insert(keyBytes, req.Data)
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

	// Delete from tree
	err = s.AVLHashTree.Delete(utils.Hash(keyBytes))
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

func (s *AVLHashTreeService) ValidateAVLHashTree(c *gin.Context) {
	if err := s.AVLHashTree.ValidateTree(); err != nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "tree_invalid",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
	})
}

func (s *AVLHashTreeService) GetAVLHashTreeProof(c *gin.Context) {
	keyParam := c.Param("key")
	if keyParam == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_key",
			Message: "Key parameter is required",
		})
		return
	}

	keyBytes, err := hex.DecodeString(keyParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_key_format",
			Message: "Key must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	proof, err := s.AVLHashTree.GenerateInclusionExclusionProof(utils.Hash(keyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "proof_generation_failed",
			Message: err.Error(),
		})
		return
	}

	publicProof := proof.ToPublicProof()
	c.JSON(http.StatusOK, convertPublicProofToResponse(publicProof, proof.ChainSize, s.convertNodeToResponse(s.AVLHashTree.Root)))
}

func (s *AVLHashTreeService) VerifyAVLHashTreeProof(c *gin.Context) {
	var req ProofVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	rootHash, err := decodeHexString(req.RootHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_root_hash",
			Message: "RootHash must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	targetKey, err := decodeHexString(req.TargetKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_target_key",
			Message: "TargetKey must be a valid hex string",
			Stack:   err.Error(),
		})
		return
	}

	hashAlgo := req.HashAlgo
	if hashAlgo == "" {
		hashAlgo = s.AVLHashTree.HashAlgo
	}

	nodes := make([]*avlhashtree.PublicCryptographicProofNode, len(req.Path))
	for i, nodeReq := range req.Path {
		keyBytes, err := decodeHexString(nodeReq.Key)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_node_key",
				Message: "Node key must be a valid hex string",
				Stack:   err.Error(),
			})
			return
		}

		dataBytes, err := decodeHexString(nodeReq.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_node_data",
				Message: "Node data must be a valid hex string",
				Stack:   err.Error(),
			})
			return
		}

		nodeHash, err := decodeHexString(nodeReq.NodeHash)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_node_hash",
				Message: "NodeHash must be a valid hex string",
				Stack:   err.Error(),
			})
			return
		}

		leftHash, err := decodeOptionalHexString(nodeReq.LeftChildSubtreeHash)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_left_subtree_hash",
				Message: "LeftChildSubtreeHash must be a valid hex string",
				Stack:   err.Error(),
			})
			return
		}

		rightHash, err := decodeOptionalHexString(nodeReq.RightChildSubtreeHash)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_right_subtree_hash",
				Message: "RightChildSubtreeHash must be a valid hex string",
				Stack:   err.Error(),
			})
			return
		}

		nodes[i] = &avlhashtree.PublicCryptographicProofNode{
			Key:                   utils.Hash(keyBytes),
			Data:                  dataBytes,
			NodeHash:              nodeHash,
			LeftChildSubtreeHash:  leftHash,
			RightChildSubtreeHash: rightHash,
		}
	}

	proof := &avlhashtree.PublicCryptographicProof{
		RootHash:  rootHash,
		TargetKey: utils.Hash(targetKey),
		Found:     req.Found,
		Path:      nodes,
		Direction: req.Direction,
		HashAlgo:  hashAlgo,
	}

	valid, err := s.AVLHashTree.VerifyPublicProof(proof)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": valid,
	})
}

func findNodeByKey(root *avlhashtree.Node, key utils.Hash) *avlhashtree.Node {
	for root != nil {
		cmp := bytes.Compare(key, root.Key)

		if cmp == 0 {
			return root
		} else if cmp < 0 {
			root = root.LeftChild
		} else {
			root = root.RightChild
		}
	}
	return nil
}

func decodeKeyToBestType(key utils.Hash) any {
	return hex.EncodeToString(key)
}

func decodeCBORToBestType(data utils.CBORData) any {
	if data == nil {
		return nil
	}

	if intData, err := utils.DecodeCBOR[int](data); err == nil {
		return intData
	}
	if strData, err := utils.DecodeCBOR[string](data); err == nil {
		return strData
	}
	if floatData, err := utils.DecodeCBOR[float64](data); err == nil {
		return floatData
	}
	if byteData, err := utils.DecodeCBOR[[]byte](data); err == nil {
		return hex.EncodeToString(byteData)
	}

	return hex.EncodeToString(data)
}

func convertPublicProofToResponse(proof *avlhashtree.PublicCryptographicProof, chainSize int, rootNode *NodeResponse) ProofResponse {
	response := ProofResponse{
		RootHash:         hex.EncodeToString(proof.RootHash),
		TargetKey:        hex.EncodeToString(proof.TargetKey),
		TargetKeyDecoded: decodeKeyToBestType(proof.TargetKey),
		Found:            proof.Found,
		Direction:        proof.Direction,
		HashAlgo:         proof.HashAlgo,
		ChainSize:        chainSize,
		RootNode:         rootNode,
	}

	if len(proof.Path) == 0 {
		return response
	}

	response.Path = make([]ProofNodeResponse, len(proof.Path))

	for i, node := range proof.Path {
		response.Path[i] = ProofNodeResponse{
			Key:                   hex.EncodeToString(node.Key),
			KeyDecoded:            decodeKeyToBestType(node.Key),
			Data:                  hex.EncodeToString(node.Data),
			DataDecoded:           decodeCBORToBestType(node.Data),
			NodeHash:              hex.EncodeToString(node.NodeHash),
			LeftChildSubtreeHash:  hex.EncodeToString(node.LeftChildSubtreeHash),
			RightChildSubtreeHash: hex.EncodeToString(node.RightChildSubtreeHash),
		}
	}

	return response
}

func decodeHexString(value string) ([]byte, error) {
	return hex.DecodeString(value)
}

func decodeOptionalHexString(value string) ([]byte, error) {
	if value == "" {
		return nil, nil
	}

	return hex.DecodeString(value)
}
