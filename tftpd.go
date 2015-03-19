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
    Packet *ReplyPacket
    Addr *net.UDPAddr
}

const BufferSize = 512

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, dispatcher *SessionDispatcher) {
    addr := net.UDPAddr {
        Port: port,
        IP: net.ParseIP(host),
    }

    conn, err := net.ListenUDP("udp", &addr)
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    sendChannel := make(chan *SendRequest)
    go Send(conn, sendChannel)

    buffer := make([]byte, BufferSize)
    for {
        _, clientAddr, _ := conn.ReadFromUDP(buffer)
        replyPacket := dispatcher.Dispatch(ConvertToUInt16(buffer[0:2]), buffer[2:], clientAddr)

        // Push sending the reply to a separate goroutine so that it doesn't block reads.
        sendRequest := SendRequest {
            Packet: replyPacket,
            Addr: clientAddr,
        }
        sendChannel <- &sendRequest
    }
}

func Send(conn *net.UDPConn, sendChannel chan *SendRequest) {
    for {
        request := <-sendChannel
        packet := *(request.Packet)
        data := packet.Write()
        _, _ = conn.WriteToUDP(data, request.Addr)
    }
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

    fmt.Printf("Listening on host %s, port %d", host, listenPort)

    var dispatcher SessionDispatcher
    Listen(*host, *listenPort, &dispatcher)
}
