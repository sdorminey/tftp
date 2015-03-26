// Connection.go implements the connection layer, which is the glue between UDP and the sessions.
// ListenForNewConnections() listens indefinitely on port 69, and when a packet comes in it spins up a
// Connection struct to communicate with the caller.
package main

import (
	"net"
	"time"
    "fmt"
)

// Represents our side of the UDP connection with the remote host.
type Connection struct {
	LastReplyPacket []byte
	Conn            *net.UDPConn
	Handler         PacketHandler
	RemoteAddr      *net.UDPAddr
    MaxRetries      int
    Timeout         time.Duration
}

// Listens for packets for the lifetime of the connection.
// - Receives packets from the remote host, and dispatches them to the session backing the connection
//   to get a reply.
// - Sends replies back to the remote host. If we haven't heard from the remote host and time out,
//   we assume our reply got lost and re-send.
// - Once the connection is done, due to success, error or timing out too much, we return and the connection
//   is destroyed.
func (c *Connection) Listen() {
	defer c.Conn.Close()

    retries := 0

	for {
		// Transmit the first reply packet of the connection, any new reply packet
		// or re-transmit a lost packet.
		if c.LastReplyPacket != nil {
			_, err := c.Conn.WriteToUDP(c.LastReplyPacket, c.RemoteAddr)
			if err != nil {
				Log.Println("Writing packet failed due to", err)
			}
		}

        // Terminate the connection if the packet handler is done with it (normally or abnormally),
        // or if we're over our retry limit.
		if c.Handler == nil || c.Handler.WantsToDie() || retries > c.MaxRetries {
			return
		}

        data, err := c.TryRead()

        // Immediately terminate the connection
        if err != nil {
            Log.Println("Error: ", err)
            return
        }

        if data != nil {
            c.LastReplyPacket = ProcessPacket(c.Handler, data)
        }
	}
}

// Tries to read a packet, timing out after a while.
// Nil is returned if there aren't bytes available.
func (c *Connection) TryRead() ([]byte, error) {
	buffer := make([]byte, MaxPacketSize)

    // Make the read attempt time out after a while so we can retry our send.
    c.Conn.SetReadDeadline(time.Now().Add(c.Timeout))
    bytesRead, clientAddr, err := c.Conn.ReadFromUDP(buffer)

    if err != nil {
        opError, isOpError := err.(*net.OpError)
        if isOpError && opError.Timeout() {
            return nil, nil
        }
        return nil, err
    }

    // Ignore requests sent to this port by other TID's.
    // Other hosts should not be able to make our connection fail.
    if !clientAddr.IP.Equal(c.RemoteAddr.IP) || clientAddr.Port != c.RemoteAddr.Port {
        return nil, nil
    }

    return buffer[:bytesRead], nil
}

// Creates a connection that will serve as our side of things.
func MakeConnection(host string, raddr *net.UDPAddr, firstPacket []byte, fs *FileSystem) (*Connection, error) {
	c := new(Connection)

	// Create a UDP listener on a random port to serve as our end of the connection.
	laddr := net.UDPAddr{
		Port: 0, // The OS will give us a random port from the ephemeral pool.
		IP:   net.ParseIP(host),
	}

	c.RemoteAddr = raddr

	conn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		return nil, err
	}
	c.Conn = conn

    handler, err := MakeHandler(firstPacket, fs)

    if err != nil {
		// No way to handle this packet, but we can send an error to
		// the remote host.
		c.LastReplyPacket = MarshalPacket(
			&ErrorPacket{
				ERR_ILLEGAL_OPERATION,
				err.Error(),
			})
	} else {
        c.Handler = handler
    }

    if c.Handler != nil {
        // Handle the first packet of information.
        c.LastReplyPacket = ProcessPacket(c.Handler, firstPacket)
    }

    // Todo: make configurable.
    c.Timeout = 3 * time.Second
    c.MaxRetries = 3

	return c, nil
}

// Creates an RRQ or WRQ handler as appropriate, to handle the packet.
// If the caller gave a bad opcode, we still need to spin up our Connection
// long enough to best-effort send an error to the caller.
func MakeHandler(packet []byte, fs *FileSystem) (PacketHandler, error) {
    if len(packet) < 2 {
        return nil, fmt.Errorf("Packet too short")
    }

    opcode := ConvertToUInt16(packet[:2])

	switch opcode {
	case PKT_RRQ:
		return MakeReadSession(fs), nil
	case PKT_WRQ:
		return MakeWriteSession(fs), nil
	default:
        return nil, fmt.Errorf("Session must start with RRQ or RWQ")
    }
}

// Listens indefinitely on the introduction port (i.e. port 69.)
// When a packet is received, a goroutine for the new connection is spun up and the
// payload of the packet is passed on to it.
func ListenForNewConnections(host string, port int, fs *FileSystem) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, MaxPacketSize)
	for {
		bytesRead, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			Log.Println("Got error listening", err)
			continue
		}

		// Create a copy so that the data won't be overwritten while it's being processed.
		data := make([]byte, bytesRead)
		copy(data, buffer[:bytesRead])

        // Now that somebody contacted us, go spin up a Connection and hand the packet we
        // received over to it for processing.
		c, err := MakeConnection(host, clientAddr, data, fs)
		if err == nil {
			go c.Listen()
		} else {
			Log.Println("Error creating connection:", err)
		}
	}
}
