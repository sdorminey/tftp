// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.

package main

import "fmt"

type SessionKiller interface {
	WantsToDie() bool // Grim!
	MakeWantToDie()
}

type PacketHandler interface {
	ProcessRead(p *ReadRequestPacket) Packet
	ProcessWrite(p *WriteRequestPacket) Packet
	ProcessData(p *DataPacket) Packet
	ProcessAck(p *AckPacket) Packet
	ProcessError(p *ErrorPacket) Packet
	SessionKiller
}

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

// Write Session (WRQ):

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
    if s.Writer.GetNumBlocks() != packet.Block - 1 {
        return MakeErrorReply(ERR_ILLEGAL_OPERATION, "Out of order")
    }

	s.Writer.Append(packet.Data)

	if len(packet.Data) < 512 {
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

func (s *WriteSession) ProcessError(packet *ErrorPacket) Packet {
	s.ShouldDie = true
	return nil
}

// Read Session (RRQ):

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

func (s *ReadSession) ProcessError(packet *ErrorPacket) Packet {
	s.ShouldDie = true
	return nil
}

func MakeDataReply(s *ReadSession) Packet {
	return &DataPacket{s.Reader.Block, s.Reader.ReadBlock()}
}

func MakeErrorReply(errCode uint16, msg string) Packet {
	return &ErrorPacket{errCode, msg}
}

// Dispatch

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

func ProcessPacket(s PacketHandler, requestPacket []byte) (marshalled []byte) {
	var reply Packet

	unmarshalled, _ := UnmarshalPacket(requestPacket)

    Log.Println("Received", unmarshalled)

    reply = Dispatch(s, unmarshalled)

	Log.Println("Sent", reply)

    if reply != nil {
        marshalled = MarshalPacket(reply)
    }

	return marshalled
}
