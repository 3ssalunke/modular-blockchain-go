package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/3ssalunke/go-blockchain/api"
	"github.com/3ssalunke/go-blockchain/core"
	"github.com/3ssalunke/go-blockchain/crypto"
)

type ServerOpts struct {
	APIListenAddr string
	ListenAddr    string
	ID            string
	RPCDecodeFunc
	RPCProcessor
	BlockTime  time.Duration
	PrivateKey *crypto.PrivateKey
}

type Server struct {
	*ServerOpts
	TCPTransport *TCPTransport
	memPool      *TxPool
	chain        *core.Blockchain
	isValidator  bool
	peerChan     chan *TCPPeer

	peerMapMU sync.RWMutex
	peerMap   map[NetAddr]*TCPPeer

	rpcCh    chan RPC
	quitChan chan struct{}
	txCh     chan *core.Transaction
}

func NewServer(opts *ServerOpts) (*Server, error) {
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

	txChan := make(chan *core.Transaction)

	if opts.APIListenAddr != "" {
		apiServerConfig := api.ServerConfig{
			ListenAddr: opts.APIListenAddr,
		}

		apiServer := api.NewServer(apiServerConfig, chain, txChan)
		go apiServer.Start()
	}

	peerChan := make(chan *TCPPeer)

	tr := NewTCPTransport(opts.ListenAddr, peerChan)

	s := &Server{
		ServerOpts:   opts,
		TCPTransport: tr,
		memPool:      NewTxPool(1000),
		chain:        chain,
		isValidator:  opts.PrivateKey != nil,
		peerChan:     peerChan,
		peerMap:      make(map[NetAddr]*TCPPeer),
		rpcCh:        make(chan RPC),
		quitChan:     make(chan struct{}, 1),
		txCh:         txChan,
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
	s.TCPTransport.Start()
free:
	for {
		select {
		case peer := <-s.peerChan:
			s.peerMapMU.Lock()
			defer s.peerMapMU.Unlock()

			fmt.Printf("new peer %+v\n", peer)

			s.peerMap[peer.conn.RemoteAddr()] = peer

			go peer.readLoop(s.rpcCh)

		case tx := <-s.txCh:
			if err := s.processTransaction(tx); err != nil {
				fmt.Println("process TX error", err)
			}

		case rpc := <-s.rpcCh:
			msg, err := s.RPCDecodeFunc(rpc)
			if err != nil {
				fmt.Println(err)
				continue
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

// func (s *Server) bootstrapNodes() {
// 	for _, tr := range s.Transports {
// 		if s.Transport.Addr() != tr.Addr() {
// 			if err := s.Transport.Connect(tr); err != nil {
// 				fmt.Println("error, could not connect to remote", err)
// 			}
// 			fmt.Println("msg, connect to remote", tr.Addr())

// 			if err := s.sendGetStatusMessage(tr); err != nil {
// 				fmt.Println("error, sendGetStatusMessage", err)
// 			}
// 		}
// 	}
// }

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
	case *GetStatusMessage:
		return s.processGetStatusMessage(msg.From, m)
	case *StatusMessage:
		return s.processStatusMessage(msg.From, m)
	case *GetBlockMessage:
		return s.processGetBlockMessage(msg.From, m)
	case *BlocksMessage:
		return s.processBlocksMessage(msg.From, m)
	default:
		return nil
	}
}

func (s *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMessage = new(GetStatusMessage)
		buf              = new(bytes.Buffer)
	)

	if err := gob.NewEncoder(buf).Encode(getStatusMessage); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) broadcast(payload []byte) error {
	// for _, tr := range s.Transports {
	// 	if err := tr.Broadcast(payload); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (s *Server) processGetStatusMessage(from NetAddr, data *GetStatusMessage) error {
	fmt.Printf("=> received status msg from %s => %+v\n", from, data)

	statusMessage := &StatusMessage{
		CurrentHeight: s.chain.Height(),
		ID:            s.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}

	s.peerMapMU.RLock()
	defer s.peerMapMU.RUnlock()

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) processStatusMessage(from NetAddr, data *StatusMessage) error {
	if data.CurrentHeight <= s.chain.Height() {
		fmt.Println("can not sync block height to low", data.CurrentHeight, s.chain.Height())
		return nil
	}

	return s.requestBlockLoop(from)
}

func (s *Server) processGetBlockMessage(from NetAddr, data *GetBlockMessage) error {
	fmt.Println("msg | received getBlocks message | from", from)

	blocks := []*core.Block{}

	if data.To == 0 {
		for i := int(data.From); i < int(s.chain.Height()); i++ {
			block, err := s.chain.GetBlockByHeight(uint32(i))
			if err != nil {
				return err
			}
			blocks = append(blocks, block)
		}
	}
	blocksMessage := &BlocksMessage{
		Blocks: blocks,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMessage); err != nil {
		return err
	}

	peer, ok := s.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", from)
	}
	msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
	return peer.Send(msg.Bytes())
}

func (s *Server) processBlocksMessage(from NetAddr, data *BlocksMessage) error {
	fmt.Println("msg | received blocks message | from", from)

	for _, block := range data.Blocks {
		if err := s.chain.AddBlock(block); err != nil {
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

	// go s.broadcastTx(tx)

	s.memPool.Add(tx)

	return nil
}

func (s *Server) requestBlockLoop(peerAddr NetAddr) error {
	ticker := time.NewTicker(3 * time.Second)

	for {
		getBlocksMessage := &GetBlockMessage{
			From: s.chain.Height() + 1,
			To:   0,
		}

		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlocksMessage); err != nil {
			return err
		}

		s.peerMapMU.RLock()
		defer s.peerMapMU.RUnlock()

		peer, ok := s.peerMap[peerAddr]
		if !ok {
			return fmt.Errorf("peer %s not known", peerAddr)
		}
		if err := peer.Send(buf.Bytes()); err != nil {
			fmt.Println("error | failed to send to peer | err |", err)
		}
		<-ticker.C
	}
}

func (s *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeBlock, buf.Bytes())

	return s.broadcast(msg.Bytes())
}

// func (s *Server) broadcastTx(tx *core.Transaction) error {
// 	buf := &bytes.Buffer{}
// 	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
// 		return err
// 	}
// 	msg := NewMessage(MessageTypeTx, buf.Bytes())

// 	return s.broadcast(msg.Bytes())
// }

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

// func (s *Server) initTransports() {
// 	for _, tr := range s.Transports {
// 		go func(tr Transport) {
// 			for rpc := range tr.Consume() {
// 				s.rpcCh <- rpc
// 			}
// 		}(tr)
// 	}
// }
