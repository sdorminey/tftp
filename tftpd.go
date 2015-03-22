// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
    "fmt";
    "flag";
    "net"
)

const BufferSize = 512

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, lifecycle *SessionLifecycle) {
    addr := net.UDPAddr {
        Port: port,
        IP: net.ParseIP(host),
    }

    conn, err := net.ListenUDP("udp", &addr)
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    buffer := make([]byte, BufferSize)
    for {
        bytesRead, clientAddr, _ := conn.ReadFromUDP(buffer)

        addr := ClientIdentity { clientAddr.IP.String(), clientAddr.Port }
        dataToSend := lifecycle.ProcessPacket(addr, buffer[:bytesRead])

        _, _ = conn.WriteToUDP(dataToSend, clientAddr) // Todo: log error
    }
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

    fmt.Printf("Listening on host %s, port %d", host, listenPort)

    lifecycle := new(SessionLifecycle)
    Listen(*host, *listenPort, lifecycle)
}
