package network

import (
	"bytes"
	"fmt"
	"time"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/3ssalunke/go-blockchain/crypto"
)

type ServerOpts struct {
	RPCDecodeFunc
	RPCProcessor
	Transports []Transport
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	ServerOpts
	memPool     *TxPool
	chain       *core.Blockchain
	isValidator bool
	rpcCh       chan RPC
	quitChan    chan struct{}
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecoderFunc
	}

	genesisBlock, err := core.GenesisBlock()
	if err != nil {
		return nil, err
	}

	chain, err := core.NewBlockchain(genesisBlock)
	if err != nil {
		return nil, err
	}
	s := &Server{
		ServerOpts:  opts,
		memPool:     NewTxPool(1000),
		chain:       chain,
		isValidator: opts.PrivateKey != nil,
		rpcCh:       make(chan RPC),
		quitChan:    make(chan struct{}, 1),
	}

	if opts.RPCProcessor == nil {
		opts.RPCProcessor = s
	}

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (s *Server) Start() {
	s.initTransports()

free:
	for {
		select {
		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				fmt.Println(err)
			}
			if err := s.RPCProcessor.ProcessMessage(msg); err != nil {
				fmt.Println(err)
			}
		case <-s.quitChan:
			break free
		}
	}

	fmt.Println("Server shutdown")
}

func (s *Server) validatorLoop() {
	ticker := time.NewTicker(s.BlockTime)
	for {
		<-ticker.C
		s.createNewBlock()
	}
}

func (s *Server) ProcessMessage(msg *DecodedMessage) error {
	switch m := msg.Data.(type) {
	case *core.Transaction:
		return s.processTransaction(m)
	case *core.Block:
		return s.processBlock(m)
	default:
		return nil
	}
}

func (s *Server) broadcast(payload []byte) error {
	for _, tr := range s.Transports {
		if err := tr.Broadcast(payload); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) processBlock(block *core.Block) error {
	if err := s.chain.AddBlock(block); err != nil {
		return err
	}

	go s.broadcastBlock(block)

	return nil
}

func (s *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	if s.memPool.Contains(hash) {
		fmt.Printf("transaction already in mempool. hash: %s", hash)
		return nil
	}

	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	fmt.Printf("adding new transaction to mempool. hash: %s", hash)

	go s.broadcastTx(tx)

	s.memPool.Add(tx)

	return nil
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

func (s *Server) createNewBlock() error {
	currentHeader, err := s.chain.GetHeader(s.chain.Height())
	if err != nil {
		return err
	}

	txx := s.memPool.Pending()

	block, err := core.NewBlockFromPrevHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	if err = block.Sign(*s.PrivateKey); err != nil {
		return err
	}

	s.memPool.ClearPending()

	go s.broadcastBlock(block)

	return s.chain.AddBlock(block)
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
