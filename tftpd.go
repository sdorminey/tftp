// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
	"flag"
	"fmt"
	"net"
    "time"
)

// A requesting host chooses its source TID as described above, and sends
// its initial request to the known TID 69 decimal (105 octal) on the
// serving host.  The response to the request, under normal operation,
// uses a TID chosen by the server as its source TID and the TID chosen
// for the previous message by the requestor as its destination TID.
func RunConnection(raddr *net.UDPConn, firstPacket []byte) {
    // Choose a TID for the server for this connection.
    laddr := net.UDPAddr {
        Port: 0, // The OS will give us a port from the ephemeral pool.
        IP: net.ParseIP("127.0.0.1"),
    }

    // Last reply packet we have for the remote host.
    var lastReplyPacket []byte
	buffer := make([]byte, 768)

    conn, err := net.ListenUDP("udp", &laddr)
    if err != nil {
        return
    }
    defer conn.Close()
    conn.SetTimeout(3 * time.Second)

    for {
        bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)

        // If we have a new packet, send it to the handler for processing.
        if !conn.IsTimeout() {
            data := buffer[:bytesRead]
            lastReplyPacket = handler.ProcessPacket(data)
        }

        // Transmit any new reply we got, or re-transmit a lost packet.
        // We may have nothing to reply with (e.g. ERROR.)
        if lastReplyPacket != nil {
            _, _ = conn.WriteToUDP(lastReplyPacket, raddr)
        }

        // Check if we still want to live.
        if handler.WantsToDie() {
            return
        }
    }
}

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, lifecycle *SessionLifecycle) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 768)
	for {
		bytesRead, clientAddr, _ := conn.ReadFromUDP(buffer)
		data := buffer[:bytesRead]
		fmt.Printf(
			"Received %d bytes from client %v: %v\n",
			bytesRead,
			clientAddr,
			data)
        go RunConnection(data)
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
