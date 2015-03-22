// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
	"flag"
	"fmt"
	"net"
)

const BufferSize = 512

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, lifecycle *SessionLifecycle) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := net.ListenUDP("udp", &addr)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, BufferSize)
	for {
		bytesRead, clientAddr, _ := conn.ReadFromUDP(buffer)
		data := buffer[:bytesRead]
		fmt.Printf(
			"Received %d bytes from client %v: %v\n",
			bytesRead,
			clientAddr,
			data)

		addr := ClientIdentity{clientAddr.IP.String(), clientAddr.Port}
		dataToSend := lifecycle.ProcessPacket(addr, data)

		_, _ = conn.WriteToUDP(dataToSend, clientAddr) // Todo: log error
	}
}

// Todo: strip out panics and use error.
func main() {
	listenPort := flag.Int("port", 69, "port to listen on.")
	host := flag.String("host", "127.0.0.1", "host address to listen on.")
	flag.Parse()

	fmt.Printf("Listening on host %s, port %d\n", *host, *listenPort)

	fs := MakeFileSystem()
	lifecycle := MakeSessionLifecycle(fs)
	Listen(*host, *listenPort, lifecycle)
}
