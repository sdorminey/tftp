package main

import (
    "net"
    "time"
    "fmt"
)

const Timeout time.Duration = 3 * time.Second

// Contains all state associated with an on-going connection.
type Connection struct {
    // RRQ or WRQ handler for the connection.
    Handler             PacketHandler
    // We store the last packet, for re-transmission in case of timeout.
    LastPacket          []byte
}

func MakeConnection(
    firstPacket []byte,
    raddr *net.UDPAddr,
    fs *FileSystem) (*Connection, error) {

    laddr := &net.UDPAddr {
        IP: net.ParseIP("127.0.0.1"),
        Port: 0, // This will select a random port from the ephemeral port-space.
    }

    listener, _ := net.ListenUDP("udp", laddr)
    var handler PacketHandler

    opcode := ConvertToUInt16(firstPacket[:2])
    switch opcode {
    case PKT_RRQ:
        handler = MakeReadSession(fs)
    case PKT_WRQ:
        handler = MakeWriteSession(fs)
    default:
        panic(nil) // Todo: fix
    }

    conn := &Connection {
        Listener: listener,
        RemoteAddr: raddr,
        Handler: handler,
    }

    go conn.Listen()

    conn.Receive(firstPacket)

    return conn, nil
}

func (c *Connection) Receive(data []byte) {
    requestPacket := UnmarshalPacket(data)
    replyPacket := Dispatch(c.Handler, requestPacket)
    fmt.Printf("replyPacket %v %T\n", replyPacket, replyPacket)
    replyBytes := MarshalPacket(replyPacket)
    fmt.Printf("replyBytes %v %T\n", replyBytes, replyBytes)
    c.LastPacket = make([]byte, len(replyBytes))
    copy(c.LastPacket, replyBytes)
}

// Listens for packets from the remote TID until the connection is terminated.
func (c *Connection) Listen(
    raddr *net.UDPAddr,
) {
    Listener            *net.UDPConn
    defer c.Listener.Close()
    buffer := make([]byte, 768)
    fmt.Println("Starting to listen")

    var err error
    for bytesRead, _, err := c.Listener.ReadFromUDP(buffer); err == nil; {
        fmt.Println("Received %d bytes.", bytesRead)
        c.TimeLastReceived.Reset(Timeout)
        c.Receive(buffer[:bytesRead])
        _, err := c.Listener.WriteToUDP(c.LastPacket, c.RemoteAddr)
        if c.Handler.WantsToDie() {
        }
    }
}

func (c *Connection) Terminate() {
    c.Listener.Close()
    _ = c.TimeLastReceived.Stop()
}
