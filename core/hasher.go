package core

import (
	"crypto/sha256"

	"github.com/3ssalunke/go-blockchain/types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct{}

func (BlockHasher) Hash(hd *Header) types.Hash {
	h := sha256.Sum256(hd.Bytes())
	return types.Hash(h)
}

type TxHasher struct{}

func (TxHasher) Hash(tx *Transaction) types.Hash {
	h := sha256.Sum256(tx.Data)
	return types.Hash(h)
}
