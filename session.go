// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.

// There are two types of sessions: ReadSession and WriteSession.
// - WriteSession is established by a WRQ packet.
//   It keeps track of the last block ID
package main

import "fmt"

type PacketHandler interface {
	ProcessRead(p *ReadRequestPacket) Packet
	ProcessWrite(p *WriteRequestPacket) Packet
	ProcessData(p *DataPacket) Packet
	ProcessAck(p *AckPacket) Packet
	ProcessError(p *ErrorPacket) Packet
}

// One session per UDP addr.
type WriteSession struct {
	Fs     *FileSystem
	Writer *File
}

func MakeWriteSession(fs *FileSystem) *WriteSession {
	return &WriteSession{Fs: fs}
}

func (s *WriteSession) ProcessRead(packet *ReadRequestPacket) Packet {
	panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Attempted RRQ in write session."})
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
	if s.Writer == nil {
		panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "DATA out of order"})
	}

	s.Writer.Append(packet.Data)

	if len(packet.Data) < 512 {
		s.Fs.Commit(s.Writer)
	}

	return &AckPacket{s.Writer.GetNumBlocks()}
}

func (s *WriteSession) ProcessAck(packet *AckPacket) Packet {
	panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Unexpected ACK"})
}

func (s *WriteSession) ProcessError(packet *ErrorPacket) Packet {
    return nil
}

type ReadSession struct {
    Fs *FileSystem
    Reader *FileReader
}

func (s *ReadSession) ProcessRead(packet *ReadRequestPacket) Packet {
    s.Reader = s.Fs.GetReader(packet.Filename)
    return MakeDataReply(s) // RRQ is acknowledged by sending DATA block 1.
}

func (s *ReadSession) ProcessWrite(packet *WriteRequestPacket) Packet {
    panic(fmt.Errorf("Shouldn't have gotten here"))
}

func (s *ReadSession) ProcessData(packet *DataPacket) Packet {
	panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Unexpected DATA"})
}

func (s *ReadSession) ProcessAck(packet *AckPacket) Packet {
    if packet.Block == s.Reader.Block {
        s.Reader.AdvanceBlock()
    } else {
        panic(fmt.Errorf("Todo: implement"))
    }

    return MakeDataReply(s)
}

func (s *ReadSession) ProcessError(packet *ErrorPacket) Packet {
    return nil
}

func MakeDataReply(s *ReadSession) Packet {
    return &DataPacket{s.Reader.Block, s.Reader.ReadBlock()}
}

func Dispatch(s PacketHandler, packet Packet) Packet {
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
		panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Unrecognized opcode."})
	}
}
