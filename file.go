// File.go defines the "file system."
// Files are simple linked lists of byte arrays - this keeps the implementation simple and lets the files
// scale up without much performance penalty.
package main

import (
	"container/list"
    "sync"
)

// Provides file creation and access.
type FileSystem struct {
	Files map[string]*File
    sync.Mutex // Guards every file creation or access. There should not be much contention.
}

func MakeFileSystem() *FileSystem {
    return &FileSystem{Files: make(map[string]*File)}
}

func (f *FileSystem) CreateFile(filename string) (*File, *ErrorPacket) {
    f.Lock()
    defer f.Unlock()

	if f.Files[filename] != nil {
		return nil, &ErrorPacket{ERR_FILE_ALREADY_EXISTS, ""}
	}

	return &File{Filename: filename}, nil
}

func (f *FileSystem) GetReader(filename string) (*FileReader, *ErrorPacket) {
    f.Lock()
    defer f.Unlock()

	if f.Files[filename] == nil {
		return nil, &ErrorPacket{ERR_FILE_NOT_FOUND, ""}
	}

	file := f.Files[filename]
	Log.Println("Began reading file", filename)
	return &FileReader{
		Block:   1,
		Current: file.Pages.Front(),
	}, nil
}

// Commits a file to the filesystem. The file must never be modified after this call is made.
func (f *FileSystem) Commit(file *File) *ErrorPacket {
    f.Lock()
    defer f.Unlock()

    if f.Files[file.Filename] != nil {
        return &ErrorPacket{ERR_FILE_ALREADY_EXISTS, ""}
    }

	f.Files[file.Filename] = file
	Log.Println("Added file", file.Filename)
    return nil
}

// Keeps track of the current block pointer and lets the reader advance forward.
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

func (r *FileReader) AtEnd() bool {
	return r.Current.Next() == nil
}

type File struct {
	Filename string

	// Each Page is a []byte chunk of the file.
	// All pages are 512 bytes except for the last one, which may be less.
	Pages list.List
}

func (f *File) Append(data []byte) {
	page := make([]byte, len(data))
	copy(page, data)
	f.Pages.PushBack(page)
}

func (f *File) GetNumBlocks() uint16 {
	return uint16(f.Pages.Len())
}
