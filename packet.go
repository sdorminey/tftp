// Formats byte arrays into packets.
package main

const (
    PKT_RRQ = 1
    PKT_WRQ = 2
    PKT_DATA = 3
    PKT_ACK = 4
    PKT_ERROR = 5
)

type RequestInfo struct
{
    Filename string
    Mode string
}

type DataPacket struct
{
    Block uint16
    Data []byte
}

func ParseRequestPacket(data []byte) (*RequestPacket) {
    filename := ExtractNullTerminatedString(data[0:])

    modeStartIndex := 1 + len(filename)
    mode := ExtractNullTerminatedString(data[modeStartIndex:])

    packet := RequestPacket {
        Filename: filename,
        Mode: mode,
    }
    return &packet
}

func ParseDataPacket(data []byte) (*DataPacket) {
    // Todo
}

func ExtractNullTerminatedString(data []byte) (string, error) {
    for index, value := range data {
        if value == 0 {
            return string(data[0:index]), nil
        }
    }

    panic()
}

func ConvertToUInt16(bytes []byte) uint16 {
    return uint16(buffer[0] << 8 | buffer[1])
}

// Formats the received packet payload into a Packet, which can be received by a session.
func FormatPacket(data []byte) {
    if len(data) <= 2 {
        panic()
    }
    opcode := ConvertToUInt16(data[0:2])
    payload := data[2:]

    switch opcode {
    case PKT_RRQ:
        fallthrough
    case PKT_WRQ:
        return ParseRequestPacket(payload)
    case PKT_DATA:
        return ParseDataPacket(payload)
    default:
        panic()
    }

    return nil
}
