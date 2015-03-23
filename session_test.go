package main

import (
	"reflect"
	"testing"
)

type RequestReply struct {
	Request Packet
	Reply   Packet
}

func TestSimpleReadWriteSession(t *testing.T) {
	harness := TestHarness{t}
	fs := MakeFileSystem()
	ws := MakeWriteSession(fs)
	rs := MakeReadSession(fs)

	test := []RequestReply{
		{&WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0}},
		{&DataPacket{1, MakePaddedBytes("hello")}, &AckPacket{1}},
		{&DataPacket{2, []byte("world!")}, &AckPacket{2}},
	}

	harness.RunExchanges(ws, test)

	test = []RequestReply{
		{
			&ReadRequestPacket{RequestPacket{"foo", "octal"}},
			&DataPacket{1, MakePaddedBytes("hello")},
		},
		{
			&AckPacket{1},
			&DataPacket{2, []byte("world!")},
		},
	}

	harness.RunExchanges(rs, test)
}

type TestHarness struct {
	t *testing.T
}

func (h *TestHarness) RunExchanges(session PacketHandler, exchanges []RequestReply) {
	for k, exchange := range exchanges {
		h.t.Log("Exchange", k)
		reply := Dispatch(session, exchange.Request)
		expectedReply := exchange.Reply

		if !reflect.DeepEqual(expectedReply, reply) {
			h.t.Fatal("Received unexpected reply.")
		}
	}
}

// Make a 512-byte array out of the text, for testing.
func MakePaddedBytes(text string) []byte {
	result := make([]byte, 512)
	copy(result, text[:])
	return result
}
