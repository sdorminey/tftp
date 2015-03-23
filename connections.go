package main

import (
    "net"
    "time"
    "fmt"
)

const Timeout time.Duration = 3 * time.Second

// Contains all state associated with an on-going connection.
type Connection struct {
    // Receives UDP packets from the remote host.
    RemoteAddr          *net.UDPAddr
    // RRQ or WRQ handler for the connection.
    Handler             PacketHandler
    // We store the last packet, for re-transmission in case of timeout.
    LastPacket          []byte
}

func MakeConnection(
    firstPacket []byte,
    raddr *net.UDPAddr,
    fs *FileSystem) (*Connection, error) {

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

    conn.Receive(firstPacket)

    return conn, nil
}

func (c *Connection) Receive(data []byte) {
    requestPacket := UnmarshalPacket(data)
    replyPacket := Dispatch(c.Handler, requestPacket)
    replyBytes := MarshalPacket(replyPacket)
    result := make([]byte, len(replyBytes))
    copy(result, replyBytes)
    c.LastPacket = result
}

func (c *Connection) Listen() error
{
    // Todo: make configurable.
    readTimeout := time.Second * 6

    laddr := &net.UDPAddr {
        IP: net.ParseIP("127.0.0.1"),
        Port: 0, // This will select a random port from the ephemeral port-space.
    }

    conn, _ := net.ListenUDP("udp", laddr)
    conn.SetDeadline(readTimeout)

    for {
        // Wait for the next packet from the remote host.
        bytesRead, _, err := conn.ReadFromUDP(buffer)
        if err == nil {
            // We have a packet to process.
            c.Receive(buffer[:bytesRead])
        }

        //
        _, err := conn.WriteToUDP(result)
        if err != nil {
            return err
        }
    }
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
