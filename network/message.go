package network

import "github.com/3ssalunke/go-blockchain/core"

type GetBlockMessage struct {
	From uint32
	To   uint32
}

type BlocksMessage struct {
	Blocks []*core.Block
}

type GetStatusMessage struct{}

type StatusMessage struct {
	ID            string
	CurrentHeight uint32
	Version       uint32
}
