package core

type Storage interface {
	Put(b *Block) error
}

type MemStore struct{}

func NewMemStore() *MemStore {
	return &MemStore{}
}

func (m *MemStore) Put(b *Block) error {
	return nil
}
