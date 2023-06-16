package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/3ssalunke/go-blockchain/core"
	"github.com/3ssalunke/go-blockchain/crypto"
	"github.com/3ssalunke/go-blockchain/network"
)

func main() {
	trLocal := network.NewLocalTransport("LOCAL")
	trRemote := network.NewLocalTransport("REMOTE")

	trLocal.Connect(trRemote)
	trRemote.Connect(trLocal)

	go func() {
		for {
			// trRemote.SendMessage(trLocal.Addr(), []byte("hello local"))
			if err := sendTransaction(trRemote, trLocal.Addr()); err != nil {
				fmt.Println(err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	opts := network.ServerOpts{
		Transports: []network.Transport{trLocal},
		BlockTime:  5 * time.Second,
	}

	server := network.NewServer(opts)
	server.Start()
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	privKey := crypto.GeneratePrivateKey()
	data := []byte(strconv.FormatInt(int64(rand.Intn(1000)), 10))
	tx := core.NewTransaction(data)
	tx.Sign(privKey)
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTypeTx, buf.Bytes())

	return tr.SendMessage(to, msg.Bytes())
}
