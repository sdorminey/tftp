package main

import "container/list"

// Files are linked lists of byte arrays.
// This design should work well for TFTP, because all writes are appends and all reads are sequential.
// If needed, we can make each page a multiple of the packet byte length.
type File struct
{
    Filename string

    // Each Page is a []byte chunk of the file.
    // All pages are 512 bytes except for the last one, which may be less.
    Pages List
}

func (f *File) Append(data []byte) {
    f.Pages.PushBack(data)
}
