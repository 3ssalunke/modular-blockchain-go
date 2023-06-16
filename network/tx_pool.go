package network

import (
	"sort"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/3ssalunke/go-blockchain/types"
)

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
}

type TxMapSorter struct {
	transactions []*core.Transaction
}

func NewTxMapSorter(txMap map[types.Hash]*core.Transaction) *TxMapSorter {
	txx := make([]*core.Transaction, len(txMap))

	i := 0
	for _, val := range txMap {
		txx[i] = val
		i++
	}

	s := &TxMapSorter{transactions: txx}

	sort.Sort(s)

	return s
}

func (s *TxMapSorter) Len() int {
	return len(s.transactions)
}

func (s *TxMapSorter) Less(i, j int) bool {
	return s.transactions[i].GetFirstSeen() < s.transactions[j].GetFirstSeen()
}

func (s *TxMapSorter) Swap(i, j int) {
	s.transactions[i], s.transactions[j] = s.transactions[j], s.transactions[i]
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (txp *TxPool) Transactions() []*core.Transaction {
	s := NewTxMapSorter(txp.transactions)
	return s.transactions
}

func (txp *TxPool) Add(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if txp.Has(hash) {
		return nil
	}

	txp.transactions[hash] = tx
	return nil
}

func (txp *TxPool) Has(hash types.Hash) bool {
	_, ok := txp.transactions[hash]
	return ok
}

func (txp *TxPool) Len() int {
	return len(txp.transactions)
}

func (txp *TxPool) Flush() {
	txp.transactions = make(map[types.Hash]*core.Transaction)
}
