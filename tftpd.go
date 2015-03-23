// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
	"flag"
	"fmt"
	"net"
    "time"
)

type Connection struct {
    LastReplyPacket []byte
    Conn *net.UDPConn
    Handler PacketHandler
    RemoteAddr *net.UDPAddr
}

func (c *Connection) Send() {
    if c.LastReplyPacket != nil {
        _, _ = c.Conn.WriteToUDP(c.LastReplyPacket, c.RemoteAddr)
    }
}

func MakeConnection(raddr *net.UDPAddr, firstPacket []byte, fs *FileSystem) (*Connection, error) {
    c := new(Connection)

    // A requesting host chooses its source TID as described above, and sends
    // its initial request to the known TID 69 decimal (105 octal) on the
    // serving host.  The response to the request, under normal operation,
    // uses a TID chosen by the server as its source TID and the TID chosen
    // for the previous message by the requestor as its destination TID.

    // Choose a TID for the server for this connection.
    laddr := net.UDPAddr {
        Port: 0, // The OS will give us a port from the ephemeral pool.
        IP: net.ParseIP("127.0.0.1"),
    }

    c.RemoteAddr = raddr

    // Last reply packet we have for the remote host.
    conn, err := net.ListenUDP("udp", &laddr)
    if err != nil {
        return nil, err
    }
    c.Conn = conn

    // Create an RRQ or WRQ handler as appropriate.
    switch ConvertToUInt16(firstPacket[:2]) {
    case PKT_RRQ:
        c.Handler = MakeReadSession(fs)
    case PKT_WRQ:
        c.Handler = MakeWriteSession(fs)
    default:
        // No way to handle this packet, but we can send an error to
        // the remote host.
        c.LastReplyPacket = MarshalPacket(
            &ErrorPacket{
                ERR_ILLEGAL_OPERATION,
                "Session must start with RRQ or RWQ",
            })
    }

    // Handle the first packet of information.
    c.LastReplyPacket = ProcessPacket(c.Handler, firstPacket)

    return c, nil
}

// Runs the connection. When done, the connection is terminated.
func (c *Connection) Listen() {
    defer c.Conn.Close()

	buffer := make([]byte, 768)

    for {
        // Transmit the first packet,
        // any new reply we got on the last loop,
        // or re-transmit a lost packet.
        c.Send()

        // Check if we still want to live.
        if c.Handler == nil || c.Handler.WantsToDie() {
            return
        }

        c.Conn.SetDeadline(time.Now().Add(3 * time.Second))
        // Todo: compare addr
        bytesRead, _, err := c.Conn.ReadFromUDP(buffer)

        // If we have a new packet, send it to the handler for processing.
        if err == nil {
            data := buffer[:bytesRead]
            c.LastReplyPacket = ProcessPacket(c.Handler, data)
        } else {
            fmt.Println("Error: ", err)
            return
        }
    }
}

// Dispatches UDP packets indefinitely to sessions.
func Listen(host string, port int, fs *FileSystem) {
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

        c, err := MakeConnection(clientAddr, data, fs)
        if err == nil {
            go c.Listen()
        } else {
            fmt.Println("Error creating connection:", err)
        }
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
