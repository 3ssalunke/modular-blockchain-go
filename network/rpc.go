package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/3ssalunke/go-blockchain/core"
)

type MessageType byte

const (
	MessageTypeTx        MessageType = 0x1
	MessageTypeBlock     MessageType = 0x2
	MessageTypeGetBlocks MessageType = 0x3
	MessageTypeStatus    MessageType = 0x4
	MessageTypeGetStatus MessageType = 0x5
	MessageTypeBlocks    MessageType = 0x6
)

type RPC struct {
	From    NetAddr
	Payload io.Reader
}

type Message struct {
	Header MessageType
	Data   []byte
}

func NewMessage(t MessageType, data []byte) *Message {
	return &Message{
		Header: t,
		Data:   data,
	}
}

func (msg *Message) Bytes() []byte {
	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(msg)
	return buf.Bytes()
}

type DecodedMessage struct {
	From NetAddr
	Data any
}

type RPCDecodeFunc func(RPC) (*DecodedMessage, error)

func DefaultRPCDecoderFunc(rpc RPC) (*DecodedMessage, error) {
	msg := Message{}
	if err := gob.NewDecoder(rpc.Payload).Decode(&msg); err != nil {
		return nil, fmt.Errorf("failed to decode message from %s: %s", rpc.From, err)
	}

	switch msg.Header {
	case MessageTypeTx:
		tx := new(core.Transaction)
		if err := tx.Decode(core.NewGobTxDecoder(bytes.NewReader(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: tx,
		}, nil
	case MessageTypeBlock:
		block := new(core.Block)
		if err := block.Decode(core.NewGobBlockDecoder(bytes.NewBuffer(msg.Data))); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: block,
		}, nil
	case MessageTypeStatus:
		statusMessage := new(StatusMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(statusMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: statusMessage,
		}, nil
	case MessageTypeGetStatus:
		return &DecodedMessage{
			From: rpc.From,
			Data: nil,
		}, nil
	case MessageTypeGetBlocks:
		getBlockMessage := new(GetBlockMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(getBlockMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: getBlockMessage,
		}, nil
	case MessageTypeBlocks:
		blocksMessage := new(BlocksMessage)
		if err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(blocksMessage); err != nil {
			return nil, err
		}
		return &DecodedMessage{
			From: rpc.From,
			Data: blocksMessage,
		}, nil
	default:
		return nil, fmt.Errorf("invalid message header %x", msg.Header)
	}
}

type RPCProcessor interface {
	ProcessMessage(*DecodedMessage) error
}
