package network

import (
	"bytes"
	"fmt"
	"net"
)

type TCPPeer struct {
	conn net.Conn
}

func (p *TCPPeer) Send(data []byte) error {
	if _, err := p.conn.Write(data); err != nil {
		return err
	}
	return nil
}

func (p *TCPPeer) readLoop(rpcCh chan RPC) {
	buf := make([]byte, 2048)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error from %+v\n", err)
			continue
		}
		msg := buf[:n]
		rpcCh <- RPC{
			From:    p.conn.RemoteAddr(),
			Payload: bytes.NewReader(msg),
		}
	}
}

type TCPTransport struct {
	listenAddr string
	listner    net.Listener
	peerChan   chan *TCPPeer
}

func NewTCPTransport(addr string, peerChan chan *TCPPeer) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
		peerChan:   peerChan,
	}
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listner.Accept()
		if err != nil {
			fmt.Printf("accept error from %+v\n", err)
			continue
		}
		peer := &TCPPeer{
			conn,
		}
		t.peerChan <- peer
		fmt.Printf("new TCP incoming connection => %+v\n", conn)
		// go t.readLoop(peer)
	}
}

func (t *TCPTransport) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listner = ln

	go t.acceptLoop()

	fmt.Println("TCP listening on port", t.listenAddr)

	return nil
}
