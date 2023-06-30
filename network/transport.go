package network

import "net"

type NetAddr net.Addr

type Transport interface {
	Consume() <-chan RPC
	Connect(Transport) error
	SendMessage(NetAddr, []byte) error
	Broadcast([]byte) error
	Addr() NetAddr
}
