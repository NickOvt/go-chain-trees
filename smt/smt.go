package smt

import "github.com/NickOvt/go-chain-trees/utils"

type SMT struct {
	HashAlgo       utils.HashAlgo
	Root           utils.Hash
	Nodes          map[string]*Node
	emptyHashCache map[int]utils.Hash
	emptyHash      utils.Hash
}

type Node struct {
	Key  utils.Hash     // in case of a non-leaf node it will be nil
	Data utils.CBORData // in case of a non-leaf node it will be nil
	Hash utils.Hash     // present on every node
	Path []byte         // nil in case of a root node
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
	t.Root = t.GetEmptyHash()
	t.Nodes = make(map[string]*Node)
	t.emptyHashCache = make(map[int]utils.Hash)

	t.BuildHashCache(t.HashAlgo)
}
