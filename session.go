// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.

// There are two types of sessions: ReadSession and WriteSession.
// - WriteSession is established by a WRQ packet.
//   It keeps track of the last block ID
package main

type PacketHandler interface {
    ProcessRead(p *ReadRequestPacket) Packet
    ProcessWrite(p *WriteRequestPacket) Packet
    ProcessData(p *DataPacket) Packet
    ProcessAck(p *AckPacket) Packet
    ProcessError(p *ErrorPacket) Packet
}

// One session per UDP addr.
type WriteSession struct {
    Fs *FileSystem
    Writer *File
}

func (s *WriteSession) ProcessRead(packet *ReadRequestPacket) Packet {
    panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Attempted RRQ in write session."})
}

func (s *WriteSession) ProcessWrite(packet *WriteRequestPacket) Packet {
    if s.Writer != nil {
        panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Attempted second WRQ."})
    }

    s.Writer = s.Fs.CreateFile(packet.Filename)
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
    panic(nil) // Don't emit an error if we receive an ERROR packet.
}

func Dispatch(s PacketHandler, packet Packet) Packet {
	switch packet.(type) {
	case *ReadRequestPacket:
        return s.ProcessRead(packet.(*ReadRequestPacket))
	case *WriteRequestPacket:
        return s.ProcessWrite(packet.(*WriteRequestPacket))
	case *DataPacket:
        return s.ProcessData(packet.(*DataPacket))
	case *AckPacket:
        return s.ProcessAck(packet.(*AckPacket))
	case *ErrorPacket:
        return s.ProcessError(packet.(*ErrorPacket))
	default:
		panic(&ErrorPacket{ERR_ILLEGAL_OPERATION, "Unrecognized opcode."})
	}
}
