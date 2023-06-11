package core

import (
	"testing"
	"time"

	"github.com/3ssalunke/go-blockchain/crypto"
	"github.com/3ssalunke/go-blockchain/types"
	"github.com/stretchr/testify/assert"
)

func randomBlock(height uint32) *Block {
	header := &Header{
		Version:       1,
		Height:        height,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     time.Now().UnixNano(),
	}
	tx := Transaction{
		Data: []byte("foo"),
	}
	return NewBlock(header, []Transaction{tx})
}

func TestSignBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.NotNil(t, b.Signature)
	assert.NotNil(t, b.Validator)
}

func TestVerifyBlock(t *testing.T) {
	privKey := crypto.GeneratePrivateKey()
	b := randomBlock(0)

	assert.Nil(t, b.Sign(privKey))
	assert.Nil(t, b.Verify())

	otherPrivKey := crypto.GeneratePrivateKey()
	b.Validator = otherPrivKey.PublicKey()

	assert.NotNil(t, b.Verify())
}
