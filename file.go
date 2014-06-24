package fileserver

import "time"

// file represents a file
type file struct {
	name    string
	modTime time.Time
	size    int64
	data    []byte
}
