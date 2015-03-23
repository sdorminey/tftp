package main

import (
	"reflect"
	"testing"
)

type TestHarness struct {
	t *testing.T
}

func (h *TestHarness) Verify (session PacketHandler, request, expectedReply Packet) {
    if session.WantsToDie() {
        h.t.Fatal("Session wanted to die")
    }

    reply := Dispatch(session, request)

    if !reflect.DeepEqual(expectedReply, reply) {
        h.t.Fatal("Received unexpected reply. Expected:", expectedReply, reflect.TypeOf(expectedReply), "actual:", reply, reflect.TypeOf(reply))
    }
}

func (h *TestHarness) VerifyDead(session PacketHandler) {
    if !session.WantsToDie() {
        panic(nil)
        h.t.Fatal("Session should have wanted to die.")
    }
}

func TestSimpleReadWriteSession(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws := MakeWriteSession(fs)
	rs := MakeReadSession(fs)

    h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
    h.Verify(ws, &DataPacket{1, MakePaddedBytes("hello")}, &AckPacket{1})
    h.Verify(ws, &DataPacket{2, []byte("world!")}, &AckPacket{2})
    h.VerifyDead(ws)

    h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo",  "octal"}}, &DataPacket{1, MakePaddedBytes("hello")})
    h.Verify(rs, &AckPacket{1}, &DataPacket{2, []byte("world!")})
    h.Verify(rs, &AckPacket{2}, nil)
    h.VerifyDead(rs)
}

// If two write sessions try to write "foo", the first one to complete will win while
// the second will receive ERR_FILE_ALREADY_EXISTS.
// We shouldn't say 'no' any sooner since we don't know if the connection will complete, or time out.
// As an optimization we could return the error as soon as we know the file is committed, but I don't think the RFC mandates that.
func TestConcurrentWritesToSameFile(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws1 := MakeWriteSession(fs)
	ws2 := MakeWriteSession(fs)

    h.Verify(ws1, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
    h.Verify(ws2, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
    h.Verify(ws1, &DataPacket{1, []byte("test")}, &AckPacket{1}) // Now ws1 has committed.
    h.VerifyDead(ws1)
    h.Verify(ws2, &DataPacket{1, []byte("test")}, &ErrorPacket{ERR_FILE_ALREADY_EXISTS, ""}) // Now ws2 is turned away.
    h.VerifyDead(ws2)
}

// If the file has already been committed we should immediately turn it away.
func TestSerialWritesToSameFile(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws1 := MakeWriteSession(fs)
	ws2 := MakeWriteSession(fs)

    h.Verify(ws1, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
    h.Verify(ws1, &DataPacket{1, []byte("test")}, &AckPacket{1}) // Now ws1 has committed.
    h.VerifyDead(ws1)
    h.Verify(ws2, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &ErrorPacket{ERR_FILE_ALREADY_EXISTS, ""}) // ws2 is turned away immediately.
    h.VerifyDead(ws2)
}

// Make a 512-byte array out of the text, for testing.
func MakePaddedBytes(text string) []byte {
	result := make([]byte, 512)
	copy(result, text[:])
	return result
}
