package fileserver

import "time"

// Options represents options for a file server.
type Options struct {
	// CheckInterval represents a check interval of
	// the file server's file change detection's process.
	CheckInterval time.Duration
	// IndexPage represents an index page path.
	IndexPage string
}

func (opts *Options) setDefaults() {
	if opts.CheckInterval == 0 {
		opts.CheckInterval = defaultCheckInterval
	}
	if opts.IndexPage == "" {
		opts.IndexPage = defaultIndexPage
	}
}
