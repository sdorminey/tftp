// Keeps track of sessions.
// A session is created by a RRQ or a WRQ packet.
package main

import "net"

type Session struct
{
}

type PacketReceiver interface
{
    // Consume a received packet and emit a writable packet in response.
    Receive(session Session) *WritablePacket
}

type SessionMap struct
{
    Channels map[net.UDPAddr]chan *Packet
}

func ReceivePackets(channel chan *Packet) {
}

func (s *SessionMap) CreateSession(addr net.UDPAddr) {
    channel := make(chan Packet)
    go ReceivePackets(channel)
}

func (s *SessionMap) SendPacket(addr net.UDPAddr, packet *Packet) {
    channel := s.Channels[addr]
    channel <- packet
}
