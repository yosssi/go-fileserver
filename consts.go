package fileserver

import "time"

// Byte sizes
const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

// Defaults for the file server.
const (
	defaultCacheSize     = 265 * MB
	defaultCheckInterval = 1 * time.Second
	defaultIndexPage     = "/index.html"
)
