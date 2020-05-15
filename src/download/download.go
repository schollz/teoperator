package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// PassThru wraps an existing io.Reader.
//
// It simply forwards the Read() call, while displaying
// the results from individual calls to it.
type PassThru struct {
	io.Reader
	total     int64 // Total # of bytes transferred
	byteLimit int64
}

// Read 'overrides' the underlying io.Reader's Read method.
// This is the one that will be called by io.Copy(). We simply
// use it to keep track of byte counts and then forward the call.
func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	if err == nil {
		pt.total += int64(n)
	}
	if pt.total > pt.byteLimit {
		err = fmt.Errorf("too many bytes")
	}

	return n, err
}

// Download a file and limit the number of bytes. If the bytes exceed,
// it will throw an error and delete the downloaded file.
func Download(u string, fname string, byteLimit int64) (err error) {
	// Get the data
	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			os.Remove(fname)
		}
	}()
	defer out.Close()

	// Wrap it with our custom io.Reader.
	src := &PassThru{Reader: resp.Body, byteLimit: byteLimit}

	_, err = io.Copy(out, src)

	return
}
