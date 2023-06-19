package network

import (
	"strconv"
	"testing"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool(100)
	assert.Equal(t, p.pending.Count(), 0)
}

func TestPoolAdd(t *testing.T) {
	p := NewTxPool(100)
	tx := core.NewTransaction([]byte("foo"))
	p.Add(tx)
	assert.Equal(t, p.pending.Count(), 1)

	_ = core.NewTransaction([]byte("foo"))
	assert.Equal(t, p.pending.Count(), 1)

	p.ClearPending()
	assert.Equal(t, p.pending.Count(), 0)
}

func TestSortTransactions(t *testing.T) {
	p := NewTxPool(100)
	txLen := 1000
	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.Itoa(i)))
		p.Add(tx)
	}

	assert.Equal(t, p.pending.Count(), 1000)
}
