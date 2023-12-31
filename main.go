package main

import (
	"bytes"
	"log"
	"net"
	"time"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/3ssalunke/go-blockchain/crypto"
	"github.com/3ssalunke/go-blockchain/network"
)

func main() {
	pk := crypto.GeneratePrivateKey()
	localNode := makeServer("localNode", &pk, ":3000", ":8080")

	go localNode.Start()

	time.Sleep(time.Second * 2)
	for i := 0; i < 10; i++ {
		tcpTester()
	}

	select {}
}

func tcpTester() {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		panic(err)
	}

	privKey := crypto.GeneratePrivateKey()
	data := []byte{0x01, 0x0a, 0x03, 0x0a, 0x0b}
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		panic(err)
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())
	_, err = conn.Write(msg.Bytes())
	if err != nil {
		panic(err)
	}
}

func makeServer(id string, pk *crypto.PrivateKey, addr string, apiListenAddr string) *network.Server {
	opts := &network.ServerOpts{
		APIListenAddr: apiListenAddr,
		ListenAddr:    addr,
		ID:            id,
		PrivateKey:    pk,
		BlockTime:     5 * time.Second,
	}
	s, err := network.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// func main() {
// 	trLocal := network.NewLocalTransport("LOCAL")
// 	trRemote := network.NewLocalTransport("REMOTE")

// 	trLocal.Connect(trRemote)
// 	trRemote.Connect(trLocal)

// 	go func() {
// 		for {
// 			if err := sendTransaction(trRemote, trLocal.Addr()); err != nil {
// 				fmt.Println(err)
// 			}
// 			time.Sleep(5 * time.Second)
// 		}
// 	}()

// 	opts := network.ServerOpts{
// 		Transport:  trLocal,
// 		Transports: []network.Transport{trLocal},
// 		BlockTime:  5 * time.Second,
// 	}

// 	server, err := network.NewServer(opts)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	server.Start()
// }

// func sendTransaction(tr network.Transport, to network.NetAddr) error {
// 	privKey := crypto.GeneratePrivateKey()
// 	data := []byte{0x01, 0x0a, 0x03, 0x0a, 0x0b}
// 	tx := core.NewTransaction(data)
// 	tx.Sign(privKey)
// 	buf := &bytes.Buffer{}
// 	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
// 		return err
// 	}
// 	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

// 	return tr.SendMessage(to, msg.Bytes())
// }
