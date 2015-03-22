// Formats byte arrays into packets.
package main

import "fmt"

// Packet opcodes:
const (
	PKT_RRQ   = 1
	PKT_WRQ   = 2
	PKT_DATA  = 3
	PKT_ACK   = 4
	PKT_ERROR = 5
)

// Packet error codes:
const (
	ERR_UNDEFINED           = iota
	ERR_FILE_NOT_FOUND      = iota
	ERR_ACCESS_VIOLATION    = iota
	ERR_DISK_FULL           = iota
	ERR_ILLEGAL_OPERATION   = iota
	ERR_FILE_ALREADY_EXISTS = iota
	ERR_NO_SUCH_USER        = iota
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

//          2 bytes    string   1 byte     string   1 byte
//          -----------------------------------------------
//   RRQ/  | 01/02 |  Filename  |   0  |    Mode    |   0  |
//   WRQ    -----------------------------------------------
type RequestPacket struct {
	Filename string
	Mode     string
}

type ReadRequestPacket struct {
	RequestPacket
}

type WriteRequestPacket struct {
	RequestPacket
}

//          2 bytes    2 bytes       n bytes
//          ---------------------------------
//   DATA  | 03    |   Block #  |    Data    |
//          ---------------------------------
type DataPacket struct {
	Block uint16
	Data  []byte
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

// Request Packet
func (p *RequestPacket) Marshal() []byte {
	result := make([]byte, len(p.Filename)+1+len(p.Mode)+1)
	copy(result, p.Filename)
	copy(result[len(p.Filename)+1:], p.Mode)
	return result
}

func (p *RequestPacket) Unmarshal(data []byte) {
	filename := ExtractNullTerminatedString(data[0:])

	modeStartIndex := 1 + len(filename)
	mode := ExtractNullTerminatedString(data[modeStartIndex:])

	p.Filename = filename
	p.Mode = mode
}

// Data Packet
func (p *DataPacket) Marshal() []byte {
	result := make([]byte, 2+len(p.Data))
	copy(result[0:2], ConvertFromUInt16(p.Block))
	copy(result[2:], p.Data)
	return result
}

func (p *DataPacket) Unmarshal(data []byte) {
	p.Block = ConvertToUInt16(data[0:2])
	p.Data = data[2:]
}

// Error Packet
func (p *ErrorPacket) Marshal() []byte {
	result := make([]byte, 2+1+len(p.ErrMsg))
	copy(result[0:2], ConvertFromUInt16(p.ErrorCode))
	copy(result[2:], []byte(p.ErrMsg[:]))
	return result
}

func (p *ErrorPacket) Unmarshal(data []byte) {
	p.ErrorCode = ConvertToUInt16(data[0:2])
	p.ErrMsg = ExtractNullTerminatedString(data[2:])
}

// Ack Packet
func (p *AckPacket) Marshal() []byte {
	return ConvertFromUInt16(p.Block)
}

func (p *AckPacket) Unmarshal(data []byte) {
	p.Block = ConvertToUInt16(data[0:2])
}

var packetTypes = map[uint16]func() Packet {
    PKT_RRQ: func() Packet { return new(ReadRequestPacket) },
    PKT_WRQ: func() Packet { return new(WriteRequestPacket) },
    PKT_DATA: func() Packet { return new(DataPacket) },
    PKT_ACK: func() Packet { return new(AckPacket) },
    PKT_ERROR: func() Packet { return new(ErrorPacket) },
}

// Factory methods
func UnmarshalPacket(data []byte) Packet {
    opcode := ConvertToUInt16(data[0:2])
    payload := data[2:]

    packet := packetTypes[opcode]()
    packet.Unmarshal(payload)

    return packet
}

func MarshalPacket(packet Packet) []byte {
    data := make([]byte, 512)
    marshalled := packet.Marshal()
    copy(data[2:], marshalled)
    copy(data[0:2], ConvertFromUInt16(GetOpcode(packet)))

    return data[:2+len(marshalled)]
}

func GetOpcode(packet Packet) uint16 {
    switch packet.(type) {
    case *ReadRequestPacket:
        return 1
    case *WriteRequestPacket:
        return 2
    case *DataPacket:
        return 3
    case *AckPacket:
        return 4
    case *ErrorPacket:
        return 5
    }

    panic(fmt.Errorf("Unrecognized packet type %v", packet))
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
	return uint16(buffer[0])<<8 | uint16(buffer[1])
}

func ConvertFromUInt16(value uint16) []byte {
	return []byte{byte((value & 0xFF00) >> 8), byte(value & 0x00FF)}
}
