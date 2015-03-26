// Packet.go defines the data structures of the TFTP protocol.
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
//   0         Not defined, see error message (if any).
//   1         File not found.
//   2         Access violation.
//   3         Disk full or allocation exceeded.
//   4         Illegal TFTP operation.
//   5         Unknown transfer ID.
//   6         File already exists.
//   7         No such user.
const (
	ERR_UNDEFINED           = iota
	ERR_FILE_NOT_FOUND      = iota
	ERR_ACCESS_VIOLATION    = iota
	ERR_DISK_FULL           = iota
	ERR_ILLEGAL_OPERATION   = iota
	ERR_FILE_ALREADY_EXISTS = iota
	ERR_NO_SUCH_USER        = iota
)

// Maximum size of a DATA packet payload. If a packet is received with len < 512,
// then that is the last data packet.
const FullDataPayloadLength = 512

// Largest byte array length possible for any packet.
const MaxPacketSize = FullDataPayloadLength + 4

// Provides methods for marshalling and unmarshalling between typed packets and byte arrays.
type Packet interface {
	Unmarshal(data []byte) error
	Marshal() []byte
	GetOpcode() uint16
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
	ErrorCode uint16

	ErrMsg string
}

// Opcodes
func (p *ReadRequestPacket) GetOpcode() uint16  { return PKT_RRQ }
func (p *WriteRequestPacket) GetOpcode() uint16 { return PKT_WRQ }
func (p *DataPacket) GetOpcode() uint16         { return PKT_DATA }
func (p *AckPacket) GetOpcode() uint16          { return PKT_ACK }
func (p *ErrorPacket) GetOpcode() uint16        { return PKT_ERROR }

// Request Packet
func (p *RequestPacket) Marshal() []byte {
	result := make([]byte, len(p.Filename)+1+len(p.Mode)+1)
	copy(result, p.Filename)
	copy(result[len(p.Filename)+1:], p.Mode)
	return result
}

func (p *RequestPacket) Unmarshal(data []byte) error {
	filename, err := ExtractNullTerminatedString(data)
	if err != nil {
		return err
	}

	mode, err := ExtractNullTerminatedString(data[1+len(filename):])
	if err != nil {
		return err
	}

	p.Filename = filename
	p.Mode = mode

	return nil
}

// Data Packet
func (p *DataPacket) Marshal() []byte {
	result := make([]byte, 2+len(p.Data))
	copy(result[:2], ConvertFromUInt16(p.Block))
	copy(result[2:], p.Data)
	return result
}

func (p *DataPacket) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("Input too small.")
	}

	p.Block = ConvertToUInt16(data[:2])
	p.Data = data[2:]

	return nil
}

// Error Packet
func (p *ErrorPacket) Marshal() []byte {
	result := make([]byte, 2+1+len(p.ErrMsg))
	copy(result[:2], ConvertFromUInt16(p.ErrorCode))
	copy(result[2:], []byte(p.ErrMsg[:]))
	return result
}

func (p *ErrorPacket) Unmarshal(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("Input too small.")
	}

	p.ErrorCode = ConvertToUInt16(data[:2])
	msg, err := ExtractNullTerminatedString(data[2:])
	p.ErrMsg = msg

	if err != nil {
		return err
	}

	return nil
}

// Ack Packet
func (p *AckPacket) Marshal() []byte {
	return ConvertFromUInt16(p.Block)
}

func (p *AckPacket) Unmarshal(data []byte) error {
	if len(data) != 2 {
		return fmt.Errorf("Input wrong size.")
	}

	p.Block = ConvertToUInt16(data)

	return nil
}

// Marshalling methods:

var packetTypes = map[uint16]func() Packet{
	PKT_RRQ:   func() Packet { return new(ReadRequestPacket) },
	PKT_WRQ:   func() Packet { return new(WriteRequestPacket) },
	PKT_DATA:  func() Packet { return new(DataPacket) },
	PKT_ACK:   func() Packet { return new(AckPacket) },
	PKT_ERROR: func() Packet { return new(ErrorPacket) },
}

func UnmarshalPacket(data []byte) (Packet, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("Input too small.")
	}

	opcode := ConvertToUInt16(data[:2])
	payload := data[2:]

	packet := packetTypes[opcode]()
	err := packet.Unmarshal(payload)
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func MarshalPacket(packet Packet) []byte {
	data := make([]byte, MaxPacketSize)
	marshalled := packet.Marshal()
	copy(data[2:], marshalled)
	copy(data[:2], ConvertFromUInt16(packet.GetOpcode()))

	return data[:2+len(marshalled)]
}

// Conversion helper methods:

func ExtractNullTerminatedString(data []byte) (string, error) {
	for index, value := range data {
		if value == 0 {
			return string(data[:index]), nil
		}
	}

	return "", fmt.Errorf("No null encountered.")
}

func ConvertToUInt16(buffer []byte) uint16 {
	return uint16(buffer[0])<<8 | uint16(buffer[1])
}

func ConvertFromUInt16(value uint16) []byte {
	return []byte{byte((value & 0xFF00) >> 8), byte(value & 0x00FF)}
}
