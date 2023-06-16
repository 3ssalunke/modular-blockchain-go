package network

import (
	"crypto"
	"fmt"
	"time"

	"github.com/3ssalunke/go-blockchain/core"
)

type ServerOpts struct {
	RPCHandler
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	memPool     *TxPool
	isValidator bool
	rpcCh       chan RPC
	quitChan    chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	s := &Server{
		memPool:     NewTxPool(),
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitChan:    make(chan struct{}, 1),
	}

	if opts.RPCHandler == nil {
		opts.RPCHandler = NewDefaultRPCHandler(s)
	}

	s.ServerOpts = opts

	return s
}

func (s *Server) Start() {
	s.initTransports()
	ticker := time.NewTicker(s.BlockTime)

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			if err := s.RPCHandler.HandleRPC(rpc); err != nil {
				fmt.Println(err)
			}
		case <-s.quitChan:
			break free
		case <-ticker.C:
			if s.isValidator {
				s.createNewBlock()
			}
		}
	}

	fmt.Println("Server shutdown")
}

func (s *Server) ProcessTransaction(from NetAddr, tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Has(hash) {
		fmt.Printf("transaction already in mempool. hash: %s", hash)
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	fmt.Printf("adding new transaction to mempool. hash: %s", hash)

	return s.memPool.Add(tx)
}

func (s *Server) createNewBlock() error {
	return nil
}

func (s *Server) initTransports() {
	for _, tr := range s.Transports {
		go func(tr Transport) {
			for rpc := range tr.Consume() {
				s.rpcCh <- rpc
			}
		}(tr)
	}
}
