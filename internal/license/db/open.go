package db

//go:generate go run ./gen/ licenses.db archive.go

import (
	"bytes"
	"compress/gzip"
	"encoding/ascii85"
	"io"
)

// Open returns a reader that produces the contents of licenses.db in this directory. This is achieved by reading
// a variable (`archive`) that has been built via code generation.
func Open() (io.ReadCloser, error) {
	decoder := ascii85.NewDecoder(bytes.NewReader(archive))
	return gzip.NewReader(decoder)
}
