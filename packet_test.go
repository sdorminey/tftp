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
		{
			[]byte{'f', 'o', 'o', 'x', 'o', 'c', 't', 'a', 'l', 0},
            nil,
            &ReadRequestPacket{},
		},
		{
			[]byte{'f', 'o', 'o', 0, 'o', 'c', 't', 'a', 'l'},
            nil,
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
		{
			[]byte{},
            nil,
			&DataPacket{},
		},
		{
			[]byte{0, 1},
            nil,
			&DataPacket{},
		},
		// Ack
		{[]byte{0x00, 0x00}, &AckPacket{0}, &AckPacket{}},
		{[]byte{0}, nil, &AckPacket{}},
		{[]byte{0xFF, 0xFE}, &AckPacket{65534}, &AckPacket{}},
		{[]byte{0xFF}, nil, &AckPacket{}},
		// Err
		{
			[]byte{0x00, 0x01, byte('h'), byte('i'), 0x00},
			&ErrorPacket{1, "hi"},
			&ErrorPacket{},
		},
		{
			[]byte{0x00, 0x01, byte('h'), byte('i')},
            nil,
			&ErrorPacket{},
		},
		{
			[]byte{0, 1},
            nil,
			&ErrorPacket{},
		},
	}

	for k, test := range tests {
		t.Logf("Executing test %d.\n", k)

		unmarshalled := test.NewPacket
        err := unmarshalled.Unmarshal(test.Data)
        // If we gave nil as the data packet, it's because we expected the test to fail.
        if test.DataPacket == nil {
            if err == nil {
                t.Fatalf("Test %d failed: Should have had error unmarshalling.", k)
            }
            continue
        } else {
            if err != nil {
                t.Fatalf("Test %d failed: Hit error unmarshalling.", k)
            }
        }

		marshalled := test.DataPacket.Marshal()

		t.Logf("Data: %v, Data Packet: %v, Marshalled: %v, Unmarshalled: %v", test.Data, test.DataPacket, marshalled, unmarshalled)

        // Compare expected data to the actual marshalled DataPacket.
		if !reflect.DeepEqual(test.Data, marshalled) {
			t.Fatalf("Test %d failed: Marshalled not equal.\n", k)
		}

        // Compare unmarshalled expected data to the expected DataPacket.
		if !reflect.DeepEqual(unmarshalled, test.DataPacket) {
			t.Fatalf("Test %d failed: Unmarshalled not equal.\n", k)
		}
	}
}
