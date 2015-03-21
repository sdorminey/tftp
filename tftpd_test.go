package main

import (
    "testing"
    "reflect"
)

type RequestReply struct {
    Request Packet
    Reply Packet
}

type MarshalTestCase struct {
    Data []byte
    DataPacket Packet
    NewPacket Packet
}

func TestPacketMarshalling(t *testing.T) {
    tests := []MarshalTestCase {
        {[]byte{0x00, 0x04, 0x00, 0x00}, &AckPacket{0}, &AckPacket{}},
        {[]byte{0x00, 0x04, 0xFF, 0xFE}, &AckPacket{65534}, &AckPacket{}},
        {
            []byte{0x00, 0x05, 0x00, 0x01, byte('h'), byte('i'), 0x00},
            &ErrorPacket{1, "hi"},
            &ErrorPacket{},
        },
    }

    for k, test := range tests {
        t.Logf("Executing test %d.\n", k)
        marshalled := test.DataPacket.Marshal()
        unmarshalled := test.NewPacket
        unmarshalled.Unmarshal(marshalled)

        t.Logf("Data: %v, Data Packet: %v, Marshalled: %v, Unmarshalled: %v", test.Data, test.DataPacket, marshalled, unmarshalled)

        if !reflect.DeepEqual(test.Data, marshalled) {
            t.Fatal("Test %d failed: Marshalled not equal.\n", k)
        }

        if !reflect.DeepEqual(unmarshalled, test.DataPacket) {
            t.Fatalf("Test %d failed: Unmarshalled not equal.\n", k)
        }
    }
}

func TestSimpleWriteSession(t *testing.T) {
    test := []RequestReply {
        {&RequestPacket{"foo", "octal"}, &AckPacket{0}},
        {&DataPacket{1, MakePaddedBytes("hello")}, &AckPacket{1}},
        {&DataPacket{2, []byte("world!"[:])}, &AckPacket{2}},
    }

    RunPacketExchangeTest(test)
}

// Make a 512-byte array out of the text, for testing.
func MakePaddedBytes(text string) []byte {
    result := make([]byte, 512)
    copy(result, text[:])
    return result
}

func RunPacketExchangeTest(exchanges []RequestReply) {
}

func main() {
    // Temp
}
