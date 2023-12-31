package core

import (
	"crypto/sha256"
	"encoding/binary"

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
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(tx.Nonce))

	data := append(buf, tx.Data...)

	h := sha256.Sum256(data)
	return types.Hash(h)
}
