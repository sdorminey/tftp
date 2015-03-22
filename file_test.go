package main

import (
	"reflect"
	"testing"
)

func TestBasicFileReadWrite(t *testing.T) {
	var file *File
	var err *ErrorPacket
	var reader *FileReader
	var data []byte

	fs := MakeFileSystem()
	file, err = fs.CreateFile("foo")
	ErrorIf(t, err != nil, "Failed to create first file")

	file.Append([]byte("hi"))
	file.Append([]byte("there"))

	fs.Commit(file)

	reader = fs.GetReader("foo")
	data = reader.ReadBlock()
	ErrorIf(t, !reflect.DeepEqual(data, []byte("hi")), "Block 0 bad")
	ErrorIf(t, !reflect.DeepEqual(data, []byte("hi")), "Block 0 bad")
	reader.AdvanceBlock()
	data = reader.ReadBlock()
	ErrorIf(t, !reflect.DeepEqual(data, []byte("there")), "Block 1 bad")
	ErrorIf(t, !reflect.DeepEqual(data, []byte("there")), "Block 1 bad")
}

func ErrorIf(t *testing.T, condition bool, msg string) {
	if condition {
		t.Errorf(msg)
	}
}
