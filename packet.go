// Formats byte arrays into packets.
package main

//import "fmt"

const (
    PKT_RRQ = 1
    PKT_WRQ = 2
    PKT_DATA = 3
    PKT_ACK = 4
    PKT_ERROR = 5
)

type PacketWriter interface {
    Marshal() []byte
}

type PacketReader interface {
    Unmarshal(data []byte)
}

type Packet interface {
    PacketReader
    PacketWriter
}

// Packet types.
type RequestPacket struct {
    Filename string
    Mode string
}

type DataPacket struct {
    Block uint16
    Data []byte
}

//          2 bytes    2 bytes
//          -------------------
//   ACK   | 04    |   Block #  |
//          --------------------
type AckPacket struct {
    Block uint16
}

//          2 bytes  2 bytes        string    1 byte
//          ----------------------------------------
//   ERROR | 05    |  ErrorCode |   ErrMsg   |   0  |
//          ----------------------------------------
type ErrorPacket struct {
    //   Value     Meaning
    //   0         Not defined, see error message (if any).
    //   1         File not found.
    //   2         Access violation.
    //   3         Disk full or allocation exceeded.
    //   4         Illegal TFTP operation.
    //   5         Unknown transfer ID.
    //   6         File already exists.
    //   7         No such user.
    ErrorCode uint16

    ErrMsg string
}

// PacketWriter implementation:
func (p *RequestPacket) Marshal() []byte {
    panic(0)
}

func (p *DataPacket) Marshal() []byte {
    panic(0)
}

func (p *AckPacket) Marshal() []byte {
    converted := ConvertFromUInt16(p.Block)
    return []byte{0x00, PKT_ACK, converted[0], converted[1]}
}

func (p *ErrorPacket) Marshal() []byte {
    result := make([]byte, 2 + 2 + 1 + len(p.ErrMsg))
    result[1] = PKT_ERROR
    copy(result[2:4], ConvertFromUInt16(p.ErrorCode))
    copy(result[4:], []byte(p.ErrMsg[:]))
    return result
}

// PacketReader implementation:
func (p *RequestPacket) Unmarshal(data []byte) {
    filename := ExtractNullTerminatedString(data[0:])

    modeStartIndex := 1 + len(filename)
    mode := ExtractNullTerminatedString(data[modeStartIndex:])

    p.Filename = filename
    p.Mode = mode
}

func (p *DataPacket) Unmarshal(data []byte) {
    p.Block = ConvertToUInt16(data[0:2])
    p.Data = data[2:]
}

func (p *AckPacket) Unmarshal(data []byte) {
    p.Block = ConvertToUInt16(data[2:4])
}

func (p *ErrorPacket) Unmarshal(data []byte) {
    p.ErrorCode = ConvertToUInt16(data[0:2])
    p.ErrMsg = string(data[2:])
}

// Conversion helper methods:
func ExtractNullTerminatedString(data []byte) string {
    for index, value := range data {
        if value == 0 {
            return string(data[0:index])
        }
    }

    panic(0)
}

func ConvertToUInt16(buffer []byte) uint16 {
    return uint16(buffer[0]) << 8 | uint16(buffer[1])
}

func ConvertFromUInt16(value uint16) []byte {
    return []byte{byte((value & 0xFF00) >> 8), byte(value & 0x00FF)}
}
