package main

import (
	"net"
	"time"
)

type UDPTransport interface {
    WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
    SetReadDeadline(t time.Time) error
    ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
    Close() error
}

type UDPListener interface {
    ListenUDP(net string, laddr *net.UDPAddr) (UDPTransport, error)
}

type Connection struct {
	LastReplyPacket []byte
	Conn            UDPTransport
	Handler         PacketHandler
	RemoteAddr      *net.UDPAddr
}

// Runs the connection. When done, the connection is terminated.
func (c *Connection) Listen() {
	defer c.Conn.Close()

	buffer := make([]byte, 768)

	for {
        Log.Println("Loop")
		// Transmit the first packet,
		// any new reply we got on the last loop,
		// or re-transmit a lost packet.
		if c.LastReplyPacket != nil {
			_, err := c.Conn.WriteToUDP(c.LastReplyPacket, c.RemoteAddr)
			if err != nil {
				Log.Println("Writing packet failed due to", err)
			} else {
                Log.Println("Successfully sent packet.")
            }
		} else {
            Log.Println("No packet to send.")
        }

		// Check if we still want to live.
		if c.Handler == nil || c.Handler.WantsToDie() {
			return
		}

		c.Conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		bytesRead, clientAddr, err := c.Conn.ReadFromUDP(buffer)


		// If we have a new packet, send it to the handler for processing.
		if err == nil {
            // Ignore requests sent to this port by other TID's.
            if !clientAddr.IP.Equal(c.RemoteAddr.IP) || clientAddr.Port != c.RemoteAddr.Port {
                continue
            }

			data := buffer[:bytesRead]
			c.LastReplyPacket = ProcessPacket(c.Handler, data)
		} else {
			Log.Println("Error: ", err)
			opError, isOpError := err.(*net.OpError)
			if !isOpError || !opError.Timeout() {
				return
			}
		}
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
	laddr := net.UDPAddr{
		Port: 0, // The OS will give us a port from the ephemeral pool.
		IP:   net.ParseIP("127.0.0.1"),
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

    if c.Handler != nil {
        // Handle the first packet of information.
        c.LastReplyPacket = ProcessPacket(c.Handler, firstPacket)
    }

	return c, nil
}

// Listens on the introduction port (i.e. port 69.)
func Listen(listener UDPListener, host string, port int, fs *FileSystem) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := listener.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 768)
	for {
        Log.Println("Waiting for messages")
		bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			Log.Println("Got error listening", err)
			continue
		}

		// Create a copy so that the data won't be overwritten while it's being processed.
		data := make([]byte, bytesRead)
		copy(data, buffer[:bytesRead])

		Log.Println("Created connection for remote host", clientAddr)

		c, err := MakeConnection(clientAddr, data, fs)
		if err == nil {
			go c.Listen()
		} else {
			Log.Println("Error creating connection:", err)
		}
	}
}
