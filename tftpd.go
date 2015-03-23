// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
	"flag"
	"fmt"
	"net"
)

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, fs *FileSystem) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := net.ListenUDP("udp", &addr)
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	buffer := make([]byte, 768)
	for {
		bytesRead, clientAddr, _ := conn.ReadFromUDP(buffer)

        _, err = MakeConnection(buffer[:bytesRead], clientAddr, fs)
	}
}

// Todo: strip out panics and use error.
func main() {
	listenPort := flag.Int("port", 69, "port to listen on.")
	host := flag.String("host", "127.0.0.1", "host address to listen on.")
	flag.Parse()

	fmt.Printf("Listening on host %s, port %d\n", *host, *listenPort)

	fs := MakeFileSystem()
	Listen(*host, *listenPort, fs)
}
