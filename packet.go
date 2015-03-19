// Formats byte arrays into packets.
package main

const (
    PKT_RRQ = 1
    PKT_WRQ = 2
    PKT_DATA = 3
    PKT_ACK = 4
    PKT_ERROR = 5
)

type RequestPacket struct {
    Filename string
    Mode string
}

type DataPacket struct {
    Block uint16
    Data []byte
}

type AckPacket struct {
    Block uint16
}

type ErrorPacket struct {
    ErrorCode uint16
    ErrMsg string
}

func ParseRequestPacket(data []byte) *RequestPacket {
    filename := ExtractNullTerminatedString(data[0:])

    modeStartIndex := 1 + len(filename)
    mode := ExtractNullTerminatedString(data[modeStartIndex:])

    packet := RequestPacket {
        Filename: filename,
        Mode: mode,
    }
    return &packet
}

func ParseDataPacket(data []byte) *DataPacket {
    packet := DataPacket {
        Block: ConvertToUInt16(data[0:2]),
        Data: data[2:],
    }
    return &packet
}

func ParseAckPacket(data []byte) *AckPacket {
    packet := AckPacket {
        Block: ConvertToUInt16(data[0:2])
    }
    return &packet
}

func ParseErrorPacket(data []byte) *ErrorPacket {
    packet := ErrorPacket {
        ErrorCode: ConvertToUInt16(data[0:2])
        ErrMsg: string(data[2:])
    }
    return &packet
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
