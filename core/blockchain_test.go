package core

import (
	"testing"

	"github.com/3ssalunke/go-blockchain/types"
	"github.com/stretchr/testify/assert"
)

func newBlockchainWithGenesis(t *testing.T) *Blockchain {
	bc, err := NewBlockchain(randomBlockWithSignature(t, 0, types.Hash{}))
	assert.Nil(t, err)
	return bc
}

func TestBlockchain(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	assert.NotNil(t, bc.validator)
	assert.Equal(t, bc.Height(), uint32(0))
}

func TestAddBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)
	prevBlockHash := getPrevBlockHash(t, bc, 1)
	block := randomBlockWithSignature(t, 1, prevBlockHash)

	assert.Nil(t, bc.AddBlock(block))
	assert.Equal(t, bc.Height(), uint32(1))
	assert.NotNil(t, bc.AddBlock(randomBlockWithSignature(t, 90, types.Hash{})))
}

func TestHasBlock(t *testing.T) {
	bc := newBlockchainWithGenesis(t)

	assert.True(t, bc.HasBlock(0))
}

func getPrevBlockHash(t *testing.T, bc *Blockchain, height uint32) types.Hash {
	prevHeader, err := bc.GetHeader(height - 1)
	assert.Nil(t, err)

	return BlockHasher{}.Hash(prevHeader)
}
