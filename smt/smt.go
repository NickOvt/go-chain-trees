package smt

import "github.com/NickOvt/go-chain-trees/utils"

var precomputedHashValueMap = make(map[utils.HashAlgo]utils.Hash)

func GetEmptyHash(hashAlgo utils.HashAlgo) utils.Hash {
	if precomputedHashValue, ok := precomputedHashValueMap[hashAlgo]; ok {
		return precomputedHashValue
	}

	hash := utils.GenerateHash(hashAlgo, []byte{})
	precomputedHashValueMap[hashAlgo] = hash
	return hash
}

var emptyHashCache = make(map[int]utils.Hash)

func GetEmptyHashForLevel(hashAlgo utils.HashAlgo, level int) utils.Hash {
	if cachedEmptyHash, ok := emptyHashCache[level]; ok {
		return cachedEmptyHash
	}

	var hash utils.Hash
	if level == 0 {
		hash = GetEmptyHash(hashAlgo)
	} else {
		prevHash := GetEmptyHashForLevel(hashAlgo, level-1)
		hash = utils.ConcatHashesAndGenerateHash(hashAlgo, prevHash, prevHash)
	}

	emptyHashCache[level] = hash
	return hash
}

func BuildHashCache(hashAlgo utils.HashAlgo) {
	hashAlgoBitCount := utils.GetHashAlgoOutputBitCount(hashAlgo)

	// Calculate and fill hash
	for level := range hashAlgoBitCount {
		GetEmptyHashForLevel(hashAlgo, level)
	}
}

type SMT struct {
	HashAlgo utils.HashAlgo
	Root     utils.Hash
}

type Node struct {
	Key  utils.Hash
	Data utils.CBORData
	Hash utils.Hash
	Path []byte
}

func NewSMT(hashAlgo utils.HashAlgo) *SMT {
	return &SMT{HashAlgo: hashAlgo}
}

func (t *SMT) Init() {
	BuildHashCache(t.HashAlgo)
}
