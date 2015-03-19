// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.
package main

import "net"

type SessionKey struct {
    Host string
    Port int
}

// One session per UDP addr.
type Session struct {
}

func (s *Session) ProcessRrq(packet *RequestPacket) *ReplyPacket {
    panic(0)
}

func (s *Session) ProcessWrq(packet *RequestPacket) *ReplyPacket {
    panic(0)
}

func (s *Session) ProcessData(packet *DataPacket) *ReplyPacket {
    panic(0)
}

func (s *Session) ProcessAck(packet *AckPacket) *ReplyPacket {
    panic(0)
}

func (s *Session) ProcessError(packet *ErrorPacket) *ReplyPacket {
    panic(0)
}

type SessionDispatcher struct {
    Session map[SessionKey]*Session
}

func (s *SessionDispatcher) Dispatch(opcode uint16, payload []byte, addr *net.UDPAddr) *ReplyPacket {
    session := s.Session[SessionKey{string(addr.IP), addr.Port}]
    if session != nil {
        panic(0)
    }

    switch opcode {
    case PKT_RRQ:
        return session.ProcessRrq(ParseRequestPacket(payload))
    case PKT_WRQ:
        return session.ProcessWrq(ParseRequestPacket(payload))
    case PKT_DATA:
        return session.ProcessData(ParseDataPacket(payload))
    case PKT_ACK:
        return session.ProcessAck(ParseAckPacket(payload))
    case PKT_ERROR:
        return session.ProcessError(ParseErrorPacket(payload))
    default:
        panic(0)
    }
}
