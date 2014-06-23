package fileserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

// FileServer is an interface for serving files.
type FileServer interface {
	Serve(root http.FileSystem) http.Handler
	Detect() (chan<- struct{}, <-chan struct{})
}

// fileServer represents a file server.
type fileServer struct {
	checkInterval time.Duration
	indexPage     string
	cache         map[string]file
}

func (fs *fileServer) Serve(root http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}

		name := path.Clean(upath)

		if file, ok := fs.cache[name]; ok {
			http.ServeContent(w, r, file.name, file.modTime, bytes.NewReader(file.data))
			// io.Copy(w, bytes.NewReader(file.data))
			return
		}

		// redirect .../index.html to .../
		// can't use Redirect() because that would make the path absolute,
		// which would be a problem running under StripPrefix
		if strings.HasSuffix(r.URL.Path, fs.indexPage) {
			localRedirect(w, r, "./")
			return
		}

		f, err := root.Open(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		d, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return
			}
		}

		// use contents of index.html for directory, if present
		if d.IsDir() {
			index := name + fs.indexPage
			ff, err := root.Open(index)
			if err == nil {
				defer ff.Close()
				dd, err := ff.Stat()
				if err == nil {
					d = dd
					f = ff
				}
			}
		}

		// Still a directory? (we didn't find an index.html file)
		if d.IsDir() {
			if checkLastModified(w, r, d.ModTime()) {
				return
			}
			dirList(w, f)
			return
		}

		http.ServeContent(w, r, d.Name(), d.ModTime(), f)

		var buf bytes.Buffer
		io.Copy(&buf, f)

		fs.cache[name] = file{
			name:    d.Name(),
			modTime: d.ModTime(),
			data:    buf.Bytes(),
		}
	})
}

func (fs *fileServer) Detect() (chan<- struct{}, <-chan struct{}) {
	quitC, doneC := make(chan struct{}), make(chan struct{})
	return quitC, doneC
}

// New creates and returns a file server.
func New(opts Options) FileServer {
	opts.setDefaults()
	return &fileServer{
		checkInterval: opts.CheckInterval,
		indexPage:     opts.IndexPage,
		cache:         make(map[string]file),
	}
}

// Quit terminates the file server's detection goroutine.
func Quit(quitC chan<- struct{}, doneC <-chan struct{}) {
	// TODO Implement Quit
	quitC <- struct{}{}
	<-doneC
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

func dirList(w http.ResponseWriter, f http.File) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for {
		dirs, err := f.Readdir(100)
		if err != nil || len(dirs) == 0 {
			break
		}
		for _, d := range dirs {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
			// name may contain '?' or '#', which must be escaped to remain
			// part of the URL path, and not indicate the start of a query
			// string or fragment.
			url := url.URL{Path: name}
			fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), htmlReplacer.Replace(name))
		}
	}
	fmt.Fprintf(w, "</pre>\n")
}
