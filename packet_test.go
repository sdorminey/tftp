package main

import (
	"reflect"
	"testing"
)

type MarshalTestCase struct {
	Data       []byte
	DataPacket Packet
	NewPacket  Packet
}

// Tests that packets are marshalled to and from binary correctly.
func TestPacketMarshalling(t *testing.T) {
	// Opcodes are omitted. Those will be tested separately.
	tests := []MarshalTestCase{
		// Read Request
		{
			[]byte{'f', 'o', 'o', 0, 'o', 'c', 't', 'a', 'l', 0},
			&ReadRequestPacket{RequestPacket{"foo", "octal"}},
			&ReadRequestPacket{},
		},
		// Write Request
		{
			[]byte{'f', 'o', 'o', 0, 'o', 'c', 't', 'a', 'l', 0},
			&WriteRequestPacket{RequestPacket{"foo", "octal"}},
			&WriteRequestPacket{},
		},
		// Data
		{
			[]byte{0xFF, 0x0F, byte('h'), byte('i')},
			&DataPacket{0xFF0F, []byte("hi")},
			&DataPacket{},
		},
		// Ack
		{[]byte{0x00, 0x00}, &AckPacket{0}, &AckPacket{}},
		{[]byte{0xFF, 0xFE}, &AckPacket{65534}, &AckPacket{}},
		// Err
		{
			[]byte{0x00, 0x01, byte('h'), byte('i'), 0x00},
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

