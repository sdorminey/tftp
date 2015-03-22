// Todo: thread safety.
package main

import "container/list"

// Accesses files.
type FileSystem struct {
	Files map[string]*File
}

func MakeFileSystem() *FileSystem {
	return &FileSystem{
		Files: make(map[string]*File),
	}
}

func (f *FileSystem) CreateFile(filename string) (*File, *ErrorPacket) {
	if f.Files[filename] != nil {
		return nil, &ErrorPacket{ERR_FILE_ALREADY_EXISTS, ""}
	}

	return &File{Filename: filename}, nil
}

func (f *FileSystem) GetReader(filename string) *FileReader {
	if f.Files[filename] == nil {
		panic(&ErrorPacket{ERR_FILE_NOT_FOUND, ""})
	}

	file := f.Files[filename]
	return &FileReader{
        Block: 1,
		Current: file.Pages.Front(),
	}
}

func (f *FileSystem) Commit(file *File) {
	f.Files[file.Filename] = file
}

type FileReader struct {
    Block   uint16
	Current *list.Element
}

func (r *FileReader) ReadBlock() []byte {
	result := r.Current.Value
	return result.([]byte)
}

func (r *FileReader) AdvanceBlock() {
	r.Current = r.Current.Next()
    r.Block++
}

// Files are linked lists of byte arrays.
// This design should work well for TFTP, because all writes are appends and all reads are sequential.
// If needed, we can make each page a multiple of the packet byte length.
type File struct {
	Filename string

	// Each Page is a []byte chunk of the file.
	// All pages are 512 bytes except for the last one, which may be less.
	Pages list.List
}

func (f *File) Append(data []byte) {
	f.Pages.PushBack(data)
}

func (f *File) GetNumBlocks() uint16 {
	return uint16(f.Pages.Len())
}
