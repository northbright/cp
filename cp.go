package cp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

var (
	DefBufSize = uint(64 * 1024)
)

// CopyFile copies src to dst and blocks the caller's goroutine until the copy is done.
// It returns the number of bytes copied.
// ctx: carries deadlines, cancellation signals.
// dst: destination file.
// src: source file.
// bufSize: buffer size.
func CopyFile(ctx context.Context, dst, src string, bufSize uint) (int64, error) {
	// Create dst dir if need.
	dstDir := filepath.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dstDir, 0755); err != nil {
		return 0, err
	}

	w, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer w.Close()

	r, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	// Start a goroutine to do IO copy.
	ch := iocopy.Start(
		// Context
		ctx,
		// Writer
		w,
		// Reader(src)
		r,
		// Buffer size
		bufSize,
		// Interval to report written bytes
		0)

	// Read the events from the channel.
	for event := range ch {
		switch ev := event.(type) {
		case *iocopy.EventStop:
			// Context is canceled or
			// context's deadline exceeded.
			return 0, ev.Err()

		case *iocopy.EventError:
			// an error occured.
			// Get the error.
			return 0, ev.Err()

		case *iocopy.EventOK:
			// IO copy succeeded.
			// Get the total count of written bytes.
			n := ev.Written()

			// Set mode of the dest file.
			info, err := os.Stat(src)
			if err != nil {
				return 0, err
			}

			if err = os.Chmod(dst, info.Mode()); err != nil {
				return 0, err
			}

			return n, nil
		}
	}

	return 0, fmt.Errorf("unknown error")
}
