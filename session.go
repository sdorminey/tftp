// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.

// There are two types of sessions: ReadSession and WriteSession.
// - WriteSession is established by a WRQ packet.
//   It keeps track of the last block ID 
package main

//import "net"

type SessionKey struct {
    Host string
    Port int
}

// One session per UDP addr.
type Session struct {
}

func (s *Session) Dispatch(opcode uint16, p Packet) *Packet {
    switch p.(type) {
    case *RequestPacket:
    case *DataPacket:
    case *AckPacket:
    case *ErrorPacket:
    default:
        panic(0)
    }
    return nil
}
