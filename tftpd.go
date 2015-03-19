// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

// Architecture:
// 
// Packets from 'net -> UDPListener -> [PacketFormatter,
//                                      SessionMap]
//
//              -> Channel for Session (guarantees serialized access to session)
//              -> Action on packet -> 

package main

import (
    "fmt";
    "flag";
    "net"
)

const BufferSize = 512

// Dispatches UDP packets indefinitely to receiver sessions.
func Listen(host string, port uint16, sendChannel chan *WritablePacket) {
    addr := net.UDPAddr {
        Port: *listenPort,
        IP: net.ParseIP(*host),
    }

    conn, err := net.ListenUDP("udp", &addr)
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    go Receive(conn)
    go Send(conn, sendChannel)
}

func Receive(conn *net.UDPConn) {
    buffer := make([]byte, BufferSize)
    for {
        bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
        packet := FormatPacket(buffer[0:bytesRead])
    }
}

func Send(conn *net.UDPConn, sendChannel chan *WritablePacket) {
    for {
        packet <- sendChannel
        data := packet.Format()
        err := conn.WriteToUDP(data, addr)
    }
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

    Listen(host, listenPort)
}
