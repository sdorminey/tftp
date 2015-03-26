// Session.go defines ReadSession and WriteSession, as well as methods for dispatching
// packets to sessions.
package main

import "fmt"

// Sessions stay alive as long as the connection hasn't completed or terminated abnormally.
type SessionKiller interface {
	// Serves as a signal to the connection layer to terminate the connection.
	WantsToDie() bool
	// Forces the session to signal its termination.
	MakeWantToDie()
}

// This interface bridges the connection layer with the session layer.
// Each method accepts one packet type, and returns one packet.
// In case of normal termination, or if an ERROR packet is received, nil is returned instead.
type PacketHandler interface {
	ProcessRead(p *ReadRequestPacket) Packet
	ProcessWrite(p *WriteRequestPacket) Packet
	ProcessData(p *DataPacket) Packet
	ProcessAck(p *AckPacket) Packet
	ProcessError(p *ErrorPacket) Packet
	SessionKiller
}

// A session contains the state of a connection.
// There are two (embedded) types of Sessions: ReadSession (for RRQ) and WriteSession (for WRQ.)
// The PacketHandler interface methods mutate the session's state, and return packets to be delivered to the remote host.
type Session struct {
	ShouldDie bool
	Fs        *FileSystem
}

func (s *Session) WantsToDie() bool {
	return s.ShouldDie
}

func (s *Session) MakeWantToDie() {
	s.ShouldDie = true
}

func (s *Session) ProcessError(packet *ErrorPacket) Packet {
	s.ShouldDie = true
	return nil
}

// Write Session (WRQ)
type WriteSession struct {
	Session
	Writer *File
}

func MakeWriteSession(fs *FileSystem) *WriteSession {
	return &WriteSession{Session{false, fs}, nil}
}

func (s *WriteSession) ProcessRead(packet *ReadRequestPacket) Packet {
	return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Bad packet")
}

func (s *WriteSession) ProcessWrite(packet *WriteRequestPacket) Packet {
	var err *ErrorPacket
	s.Writer, err = s.Fs.CreateFile(packet.Filename)
	if err != nil {
		return err
	}
	return &AckPacket{0}
}

func (s *WriteSession) ProcessData(packet *DataPacket) Packet {
	// Ignore duplicated DATA packets.
	if s.Writer.GetNumBlocks() >= packet.Block {
		return nil
	}

	// Packets from the future cannot be explained except with a time machine,
	// since acknowledgement is lock-step and they should have gotten an ACK for block b before
	// sending DATA for block n > b.
	if s.Writer.GetNumBlocks() != packet.Block-1 {
		return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Out of order")
	}

	s.Writer.Append(packet.Data)

	// If a DATA packet is less than the maximum length, then it must be the last packet.
	if len(packet.Data) < FullDataPayloadLength {
		// We may fail to commit if another write session won a race to write the same file.
		// But whether successful or unsuccessful, we should die now.
		err := s.Fs.Commit(s.Writer)
		s.ShouldDie = true
		if err != nil {
			return err
		}
	}

	return &AckPacket{s.Writer.GetNumBlocks()}
}

func (s *WriteSession) ProcessAck(packet *AckPacket) Packet {
	return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Bad packet")
}

// Read Session (RRQ)
type ReadSession struct {
	Session
	Reader *FileReader
}

func MakeReadSession(fs *FileSystem) *ReadSession {
	return &ReadSession{Session{false, fs}, nil}
}

func (s *ReadSession) ProcessRead(packet *ReadRequestPacket) Packet {
	reader, err := s.Fs.GetReader(packet.Filename)
	if err != nil {
		return err
	}

	s.Reader = reader
	return MakeDataReply(s) // RRQ is acknowledged by sending DATA block 1.
}

func (s *ReadSession) ProcessWrite(packet *WriteRequestPacket) Packet {
	return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Bad packet")
}

func (s *ReadSession) ProcessData(packet *DataPacket) Packet {
	return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Bad packet")
}

func (s *ReadSession) ProcessAck(packet *AckPacket) Packet {
	// Duplicate or outdated ACKs can be explained by the network, and shouldn't cause error.
	if packet.Block < s.Reader.Block {
		return nil
	}

	// Due to lock-step, this condition is impossible if the remote host is following the protocol.
	if packet.Block > s.Reader.Block {
		return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Out of order")
	}

	// Client has acknowledged the last block with an ACK.
	// Now we can die happily.
	if s.Reader.AtEnd() {
		s.ShouldDie = true
		return nil
	}

	s.Reader.AdvanceBlock()

	return MakeDataReply(s)
}

func MakeDataReply(s *ReadSession) Packet {
	return &DataPacket{s.Reader.Block, s.Reader.ReadBlock()}
}

func MakeErrorReply(errCode uint16, msg string) Packet {
	return &ErrorPacket{errCode, msg}
}

// Dispatch methods:

func DispatchInner(s PacketHandler, packet Packet) Packet {
	switch p := packet.(type) {
	case *ReadRequestPacket:
		return s.ProcessRead(p)
	case *WriteRequestPacket:
		return s.ProcessWrite(p)
	case *DataPacket:
		return s.ProcessData(p)
	case *AckPacket:
		return s.ProcessAck(p)
	case *ErrorPacket:
		return s.ProcessError(p)
	default:
		panic(fmt.Errorf("Unknown packet type."))
	}
}

// Given a packet, calls the appropriate method on the PacketHandler and returns the reply.
func Dispatch(s PacketHandler, packet Packet) Packet {
	var reply Packet

	if packet == nil {
		reply = MakeErrorReply(ERR_ILLEGAL_OPERATION, "Error parsing packet")
	} else {
		reply = DispatchInner(s, packet)
	}

	// All ERROR responses destroy the session.
	_, isError := reply.(*ErrorPacket)
	if isError {
		s.MakeWantToDie()
	}

	return reply
}

// Given raw request packet data, returns raw reply data (or nil if no response is given.)
func ProcessPacket(s PacketHandler, requestPacket []byte) (marshalled []byte) {
	unmarshalled, _ := UnmarshalPacket(requestPacket)

	Log.Println("Received", unmarshalled)

	reply := Dispatch(s, unmarshalled)

	Log.Println("Sent", reply)

	if reply != nil {
		marshalled = MarshalPacket(reply)
	}

	return marshalled
}
