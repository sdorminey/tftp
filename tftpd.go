// TFTP Daemon
// Implements RFC 1350, in octet mode only, over UDP and with files stored in memory only.

package main

import (
    "fmt";
    "flag";
    "net"
)

const BufferSize = 512

const (
    PKT_RRQ = 1
    PKT_WRQ = 2
    PKT_DATA = 3
    PKT_ACK = 4
    PKT_ERROR = 5
)

type RequestPacket struct
{
    Filename string
    Mode string
}

func ParseRequestPacket(data []byte) (*RequestPacket, error) {
    filename, err := ExtractNullTerminatedString(data[0:])
    if err != nil {
        return nil, fmt.Errorf("Trouble parsing filename")
    }

    modeStartIndex := 1 + len(filename)
    mode, err := ExtractNullTerminatedString(data[modeStartIndex:])
    if err != nil {
        return nil, fmt.Errorf("Trouble parsing mode")
    }

    packet := RequestPacket {
        Filename: filename,
        Mode: mode,
    }
    return &packet, nil
}

func ExtractNullTerminatedString(data []byte) (string, error) {
    for index, value := range data {
        if value == 0 {
            return string(data[0:index]), nil
        }
    }

    return "", fmt.Errorf("No null!")
}

func Dispatch(opcode uint16, data []byte) (error) {
    switch opcode {
    case PKT_RRQ:
        fallthrough
    case PKT_WRQ:
        packet, err := ParseRequestPacket(data)
        if err != nil {
            return err
        }
        fmt.Println(packet)
    default:
        return fmt.Errorf("Unrecognized opcode %d", opcode)
    }

    return nil
}

func main() {
    listenPort := flag.Int("port", 69, "port to listen on.")
    host := flag.String("host", "127.0.0.1", "host address to listen on.")
    flag.Parse()

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
        // Todo: validate that len(buffer) > 2.
        // Todo: use panic/recover to simplify error handling for packet transmission errors.
        opcode := uint16(buffer[0] << 8 | buffer[1])
        err = Dispatch(opcode, buffer[2:bytesRead])
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
    }
}
