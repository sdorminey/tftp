// Tests simple and complex scenarios for the sessions.

package main

import (
	"reflect"
	"testing"
)

type TestHarness struct {
	t *testing.T
}

func (h *TestHarness) Verify(session PacketHandler, request, expectedReply Packet) {
	h.t.Log("Request:", request, "expected reply:", expectedReply)
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
		h.t.Fatal("Session should have wanted to die.")
	}
}

// Happy-path test.
func TestSimpleReadWriteSession(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws := MakeWriteSession(fs)
	rs := MakeReadSession(fs)

	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
	h.Verify(ws, &DataPacket{1, MakePaddedBytes("hello")}, &AckPacket{1})
	h.Verify(ws, &DataPacket{2, []byte("world!")}, &AckPacket{2})
	h.VerifyDead(ws)

	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &DataPacket{1, MakePaddedBytes("hello")})
	h.Verify(rs, &AckPacket{1}, &DataPacket{2, []byte("world!")})
	h.Verify(rs, &AckPacket{2}, nil)
	h.VerifyDead(rs)
}

// Tries to catch the boundary condition of single-block files.
func TestSmallFile(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws := MakeWriteSession(fs)
	rs := MakeReadSession(fs)

	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
	h.Verify(ws, &DataPacket{1, []byte("world!")}, &AckPacket{1})
	h.VerifyDead(ws)

	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &DataPacket{1, []byte("world!")})
	h.Verify(rs, &AckPacket{1}, nil)
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

// RFC: All  packets other than duplicate ACK's and those used for termination are acknowledged unless a timeout occurs [4].
func TestDuplicatePacket(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws := MakeWriteSession(fs)
	rs := MakeReadSession(fs)

	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0}) // Duplicate RRQ/WRQ shouldn't happen since those go to port 69.
	h.Verify(ws, &DataPacket{1, MakePaddedBytes("hi")}, &AckPacket{1})              // If it's < 512 bytes, the session will be dead and we won't re-transmit.
	h.Verify(ws, &DataPacket{1, MakePaddedBytes("hi")}, nil)                        // We'll re-transmit the last packet (which must have been an ACK.)
	h.Verify(ws, &DataPacket{2, []byte("there")}, &AckPacket{2})                    // Duplicates of the last ACK packet can't come because we've spun down the listener.
	h.VerifyDead(ws)

	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &DataPacket{1, MakePaddedBytes("hi")})
	h.Verify(rs, &AckPacket{1}, &DataPacket{2, []byte("there")})
	h.Verify(rs, &AckPacket{1}, nil) // Don't acknowledge duplicate ACKs.
	h.Verify(rs, &AckPacket{2}, nil)
	h.VerifyDead(rs)
}

func TestInterleavedReadsAndWrites(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()
	ws1 := MakeWriteSession(fs)
	ws2 := MakeWriteSession(fs)
	rs1 := MakeReadSession(fs)
	rs2 := MakeReadSession(fs)

	// ws1 and ws2 will interleaved-ly write foo1 and foo2.
	// ws1 will commit, and then rs1 and rs2 will both read foo1.

	h.Verify(ws1, &WriteRequestPacket{RequestPacket{"foo1", "octal"}}, &AckPacket{0})
	h.Verify(ws1, &DataPacket{1, MakePaddedBytes("hi")}, &AckPacket{1})
	h.Verify(ws2, &WriteRequestPacket{RequestPacket{"foo2", "octal"}}, &AckPacket{0})
	h.Verify(ws2, &DataPacket{1, MakePaddedBytes("hi")}, &AckPacket{1})
	h.Verify(ws1, &DataPacket{2, []byte("there")}, &AckPacket{2}) // Now ws1 has committed.
	h.VerifyDead(ws1)
	h.Verify(rs1, &ReadRequestPacket{RequestPacket{"foo1", "octal"}}, &DataPacket{1, MakePaddedBytes("hi")})
	h.Verify(rs2, &ReadRequestPacket{RequestPacket{"foo1", "octal"}}, &DataPacket{1, MakePaddedBytes("hi")})
	h.Verify(ws2, &DataPacket{2, []byte("there")}, &AckPacket{2}) // Now ws2 has committed.
	h.VerifyDead(ws2)
	h.Verify(rs1, &AckPacket{1}, &DataPacket{2, []byte("there")})
	h.Verify(rs1, &AckPacket{2}, nil)
	h.VerifyDead(rs1)
	h.Verify(rs2, &AckPacket{1}, &DataPacket{2, []byte("there")})
	h.Verify(rs2, &AckPacket{2}, nil)
	h.VerifyDead(rs2)
}

// Errors are caused by three types of events: not being able to satisfy the
// request (e.g., file not found, access violation, or no such user),
// receiving a packet which cannot be explained by a delay or
// duplication in the network (e.g., an incorrectly formed packet), and
// losing access to a necessary resource (e.g., disk full or access
// denied during a transfer).
// Make a 512-byte array out of the text, for testing.
func TestErrorConditions(t *testing.T) {
	h := TestHarness{t}
	fs := MakeFileSystem()

	// Type 1: Unable to satisfy request.

	// File not found.
	rs := MakeReadSession(fs)
	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &ErrorPacket{ERR_FILE_NOT_FOUND, ""})
	h.VerifyDead(rs)

	// Type 2: Receiving packet which cannot be explained.

	// Bad packet (if marshalling fails, dispatcher gets nil.)
	rs = MakeReadSession(fs)
	h.Verify(rs, nil, &ErrorPacket{ERR_ILLEGAL_OPERATION, "Error parsing packet"})
	h.VerifyDead(rs)

	// Out-of-order packets: WRQ
	ws := MakeWriteSession(fs)
	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
	h.Verify(ws, &DataPacket{2, MakePaddedBytes("hi")}, &ErrorPacket{ERR_ILLEGAL_OPERATION, "Out of order"})
	h.VerifyDead(ws)

	// Add 'foo' to the filesystem for the next test.
	ws = MakeWriteSession(fs)
	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo", "octal"}}, &AckPacket{0})
	h.Verify(ws, &DataPacket{1, MakePaddedBytes("hello")}, &AckPacket{1})
	h.Verify(ws, &DataPacket{2, []byte("world!")}, &AckPacket{2})
	h.VerifyDead(ws)

	// Out-of-order packets: RRQ
	rs = MakeReadSession(fs)
	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &DataPacket{1, MakePaddedBytes("hello")})
	h.Verify(rs, &AckPacket{2}, &ErrorPacket{ERR_ILLEGAL_OPERATION, "Out of order"})
	h.VerifyDead(rs)

	// Wrong type of packet: RRQ
	rs = MakeReadSession(fs)
	h.Verify(rs, &ReadRequestPacket{RequestPacket{"foo", "octal"}}, &DataPacket{1, MakePaddedBytes("hello")})
	h.Verify(rs, &DataPacket{1, nil}, &ErrorPacket{ERR_ILLEGAL_OPERATION, "Bad packet"})
	h.VerifyDead(rs)

	// Wrong type of packet: WRQ
	ws = MakeWriteSession(fs)
	h.Verify(ws, &WriteRequestPacket{RequestPacket{"foo2", "octal"}}, &AckPacket{0})
	h.Verify(ws, &AckPacket{0}, &ErrorPacket{ERR_ILLEGAL_OPERATION, "Bad packet"})
	h.VerifyDead(ws)
}

func MakePaddedBytes(text string) []byte {
	result := make([]byte, 512)
	copy(result, text[:])
	return result
}
