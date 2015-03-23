package main

import (
    "net"
    "time"
)

const Timeout time.Duration = 3 * time.Second

// Contains all state associated with an on-going connection.
type Connection struct {
    // Receives UDP packets from the remote host.
    Listener            *net.UDPConn
    // Transmits UDP packets to the remote host.
    Dialer              *net.UDPConn
    // RRQ or WRQ handler for the connection.
    Handler             PacketHandler
    // We store the last packet, for re-transmission in case of timeout.
    LastPacket          []byte
    // Wakes up the Send() goroutine if it's time to retransmit.
    TimeLastReceived    *time.Timer
}

func MakeConnection(
    firstPacket []byte,
    raddr *net.UDPAddr) (*Connection, error) {

    laddr := &net.UDPAddr {
        IP: net.ParseIP("127.0.0.1"),
        Port: 0, // This will select a random port from the ephemeral port-space.
    }

    listener, _ := net.ListenUDP("udp", laddr)
    dialer, _ := net.DialUDP("udp", laddr, raddr)

    conn := &Connection {
        Listener: listener,
        Dialer: dialer,
        TimeLastReceived: time.NewTimer(Timeout),
    }

    conn.Receive(firstPacket)

    return conn, nil
}

func (c *Connection) Receive(data []byte) {
    opcode := ConvertToUInt16(data[:2])

    switch opcode {
    case PKT_RRQ:
        if c.Handler == nil {
            c.Handler = &ReadSession{}
        }
    case PKT_WRQ:
        if c.Handler == nil {
            c.Handler = &WriteSession{}
        }
    default:
        if c.Handler == nil {
            panic(nil) // Todo: not implemented.
        }
    }

    requestPacket := UnmarshalPacket(data)
    replyPacket := Dispatch(c.Handler, requestPacket)
    replyBytes := replyPacket.Marshal()
    c.LastPacket = make([]byte, len(replyBytes))
    copy(c.LastPacket, replyBytes)
}

// Listens for packets from the remote TID until the connection is terminated.
func (c *Connection) Listen() {
    buffer := make([]byte, 768)

    for bytesRead, _, err := c.Listener.ReadFromUDP(buffer); err == nil; {
        c.TimeLastReceived.Reset(Timeout)
        c.Receive(buffer[:bytesRead])
    }
}

// Dials the remote host and transfers the last packet we have for her.
// We'll retry if we haven't received any packets in long enough.
func (c *Connection) Dial() {
    // Todo: handle errors.
    for _, err := c.Dialer.Write(c.LastPacket); err == nil; {
        // Retry sending the last packet if we've been woken up from the timer.
        <-c.TimeLastReceived.C
    }
}

func (c *Connection) Terminate() {
    c.Listener.Close()
    c.Dialer.Close()
    _ = c.TimeLastReceived.Stop()
}
