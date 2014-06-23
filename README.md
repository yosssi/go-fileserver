# go-fileserver - Go cached file server

## Overview

This is a Go file server equipped with caching feature.

## Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/yosssi/go-fileserver"
)

func main() {
	fs := fileserver.New(fileserver.Options{})
	http.Handle("/", fs.Serve(http.Dir("/Users/yoshidakeiji/www")))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
```
