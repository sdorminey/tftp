// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
    "fmt";
    "flag";
    "net"
)

const BufferSize = 512

func Listen(host string, port uint16) {
    addr := net.UDPAddr {
        Port: *listenPort,
        IP: net.ParseIP(*host),
    }
    fmt.Printf("Server listening on %v\n", addr)

    conn, err := net.ListenUDP("udp", &addr)
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    buffer := make([]byte, BufferSize)
    for {
        bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
        fmt.Printf("Received %d bytes from addr %v and error %v.\n", bytesRead, clientAddr, err)
        err = Dispatch(opcode, buffer)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

    Listen(host, listenPort)
}
