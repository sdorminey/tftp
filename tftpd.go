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

type SendRequest struct {
    Packet *Packet
    Addr *net.UDPAddr
}

const BufferSize = 512

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port uint16, dispatcher SessionDispatcher) {
    addr := net.UDPAddr {
        Port: *listenPort,
        IP: net.ParseIP(*host),
    }

    conn, err := net.ListenUDP("udp", &addr)
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    sendChannel := make(chan *WritablePacket)
    go Send(conn, sendChannel)

    buffer := make([]byte, BufferSize)
    for {
        bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
        opcode, packet := FormatPacket(buffer[0:bytesRead])

        replyPacket := dispatcher.Dispatch(opcode, packet)

        // Push sending the reply to a separate goroutine so that it doesn't block reads.
        sendChannel <- SendRequest {
            Packet: replyPacket,
            Addr: clientAddr,
        }
    }
}

func Send(conn *net.UDPConn, sendChannel chan *SendRequest) {
    for {
        request <- sendChannel
        data := request.Packet.Format()
        err := conn.WriteToUDP(data, request.Addr)
    }
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

    SessionDispatcher dispatcher
    Listen(host, listenPort, dispatcher)
}
