package network

// func TestConnect(t *testing.T) {
// 	tra := NewLocalTransport()
// 	trb := NewLocalTransport(net.Addr)

// 	tra.Connect(trb)
// 	trb.Connect(tra)

// 	assert.Equal(t, tra.peers[trb.addr], trb)
// 	assert.Equal(t, trb.peers[tra.addr], tra)
// }

// func TestMessage(t *testing.T) {
// 	tra := NewLocalTransport("a")
// 	trb := NewLocalTransport("b")

// 	tra.Connect(trb)
// 	trb.Connect(tra)

// 	msg := []byte("Hello World")

// 	assert.Nil(t, tra.SendMessage(trb.addr, msg))

// 	rpc := <-trb.Consume()
// 	assert.Equal(t, rpc.Payload, msg)
// 	assert.Equal(t, rpc.From, tra.addr)
// }
