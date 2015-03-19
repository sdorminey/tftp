package main

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
    packet := DataPacket {
        Block: data[
    }
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

// Dispatches packets.
func Dispatch(data []byte) (error) {
    if len(data) < 2 {
        return fmt.Errorf("Not enough data.")
    }
    opcode := ConvertToUInt16(data[0:2])

    switch opcode {
    case PKT_RRQ:
        fallthrough
    case PKT_WRQ:
        packet, err := ParseRequestPacket(data[2:])
        if err != nil {
            return err
        }
        fmt.Println(packet)
    default:
        return fmt.Errorf("Unrecognized opcode %d", opcode)
    }

    return nil
}
