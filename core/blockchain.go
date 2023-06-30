package core

import (
	"fmt"
	"sync"
)

type Blockchain struct {
	store         Storage
	lock          sync.RWMutex
	headers       []*Header
	blocks        []*Block
	validator     Validator
	contractState *State
}

func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:       []*Header{},
		store:         NewMemStore(),
		contractState: NewState(),
	}
	bc.validator = NewBlockValidator(bc)
	err := bc.addBlockChainWithoutValidation(genesis)

	return bc, err
}

func (bc *Blockchain) AddBlock(b *Block) error {
	if err := bc.validator.ValidateBlock(b); err != nil {
		return err
	}

	for _, tx := range b.Transactions {
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
	}

	return bc.addBlockChainWithoutValidation(b)
}

func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) is too high", height)
	}

	return bc.blocks[height], nil
}

func (bc *Blockchain) GetHeader(height uint32) (*Header, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	if height > bc.Height() {
		return nil, fmt.Errorf("given height (%d) is too high", height)
	}

	return bc.headers[height], nil
}

func (bc *Blockchain) HasBlock(height uint32) bool {
	return height <= bc.Height()
}

func (bc *Blockchain) Height() uint32 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint32(len(bc.headers) - 1)
}

func (bc *Blockchain) addBlockChainWithoutValidation(b *Block) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	bc.headers = append(bc.headers, b.Header)
	bc.blocks = append(bc.blocks, b)
	return bc.store.Put(b)
}
