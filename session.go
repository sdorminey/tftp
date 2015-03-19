// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.
package main

import "net"

type Session struct
{
}

func (s *Session) ProcessRrqPacket(packet *RequestPacket) {
}

func (s *Session) ProcessWrqPacket(packet *RequestPacket) {
}

func (s *Session) ProcessDataPacket(packet *DataPacket) {
}

type SessionMap struct
{
    Sessions map[net.UDPAddr]Session
}

func (s *SessionMap) GetSession(addr net.UDPAddr) Session {
    return s.Sessions[addr]
}
