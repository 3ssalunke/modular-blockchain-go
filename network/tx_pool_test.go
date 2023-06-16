package network

import (
	"strconv"
	"testing"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/stretchr/testify/assert"
)

func TestTxPool(t *testing.T) {
	p := NewTxPool()
	assert.Equal(t, p.Len(), 0)
}

func TestPoolAdd(t *testing.T) {
	p := NewTxPool()
	tx := core.NewTransaction([]byte("foo"))
	assert.Nil(t, p.Add(tx))
	assert.Equal(t, p.Len(), 1)

	_ = core.NewTransaction([]byte("foo"))
	assert.Equal(t, p.Len(), 1)

	p.Flush()
	assert.Equal(t, p.Len(), 0)
}

func TestSortTransactions(t *testing.T) {
	p := NewTxPool()
	txLen := 1000
	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.Itoa(i)))
		assert.Nil(t, p.Add(tx))
	}

	assert.Equal(t, p.Len(), 1000)
}
