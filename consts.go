package fileserver

import "time"

// Defaults for the file server.
const (
	defaultCheckInterval = 1 * time.Second
	defaultIndexPage     = "/index.html"
)
