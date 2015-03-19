// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.
package main

import "net"

// One session per UDP addr.
type Session struct
{
}

func (s *Session) ProcessRrq(packet RequestPacket) *Packet {
    panic()
}

func (s *Session) ProcessWrq(packet RequestPacket) *Packet {
    panic()
}

func (s *Session) ProcessData(packet DataPacket) *Packet {
    panic()
}

func (s *Session) ProcessAck(packet AckPacket) *Packet {
    panic()
}

func (s *Session) ProcessError(packet ErrorPacket) *Packet {
    panic()
}

type SessionDispatcher struct
{
    Session map[net.UDPAddr]Session
}

func (s *SessionDispatcher) Dispatch(uint16 opcode, []byte payload, addr net.UDPAddr) *Packet {
    session := s.Session[addr]
    if session != nil {
        panic()
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
        panic()
    }
}
