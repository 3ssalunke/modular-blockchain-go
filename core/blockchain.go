package core

import (
	"fmt"
	"sync"

	"github.com/3ssalunke/go-blockchain/types"
)

type Blockchain struct {
	store         Storage
	lock          sync.RWMutex
	headers       []*Header
	blocks        []*Block
	blockstore    map[types.Hash]*Block
	txstore       map[types.Hash]*Transaction
	validator     Validator
	contractState *State
}

func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:       []*Header{},
		store:         NewMemStore(),
		contractState: NewState(),
		blockstore:    make(map[types.Hash]*Block),
		txstore:       make(map[types.Hash]*Transaction),
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

func (bc *Blockchain) GetBlockByHash(hash types.Hash) (*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	block, ok := bc.blockstore[hash]
	if !ok {
		return nil, fmt.Errorf("block not found for hash %s", hash)
	}

	return block, nil
}

func (bc *Blockchain) GetBlockByHeight(height uint32) (*Block, error) {
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

func (bc *Blockchain) GetTxByHash(hash types.Hash) (*Transaction, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()

	tx, ok := bc.txstore[hash]
	if !ok {
		return nil, fmt.Errorf("transaction not found for given hash")
	}

	return tx, nil
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
	bc.blockstore[b.Hash(BlockHasher{})] = b

	for _, tx := range b.Transactions {
		bc.txstore[tx.Hash(TxHasher{})] = tx
	}

	return bc.store.Put(b)
}
