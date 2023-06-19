package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	data := []byte{0x01, 0x0a, 0x03, 0x0a, 0x0b}
	vm := NewVM(data)

	assert.Nil(t, vm.Run())
	assert.Equal(t, 4, vm.stack.Pop())

	data = []byte{0x07, 0x0a, 0x03, 0x0a, 0x0e}
	vm = NewVM(data)

	assert.Nil(t, vm.Run())
	assert.Equal(t, 4, vm.stack.Pop())
}
