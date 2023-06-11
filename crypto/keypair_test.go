package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPairSignVerify(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	// address := pubKey.Address()

	msg := []byte("hello world")
	sig, err := privKey.Sign(msg)
	assert.Nil(t, err)
	assert.True(t, sig.Verify(pubKey, msg))

	otherPrivKey := GeneratePrivateKey()
	otherPubKey := otherPrivKey.PublicKey()
	assert.False(t, sig.Verify(otherPubKey, msg))

	otherMsg := []byte("hello world again")
	assert.False(t, sig.Verify(pubKey, otherMsg))
}
