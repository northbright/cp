package cp

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/iocopy/progress"
	"github.com/northbright/pathelper"
)

var (
	// ErrNotFSRegularFile represents the error that src is not a regular file in the file system.
	ErrNotFSRegularFile = errors.New("not a regular file in file system")
)

// CopyFile copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// srcFS: file system of src.
// src: source file path in the file system.
// dst: destination file.
// options: [progress.Option] used to report progress.
func CopyFSFile(ctx context.Context, srcFS fs.FS, src, dst string, options ...progress.Option) (written int64, err error) {
	// Open the src file.
	fSrc, err := srcFS.Open(src)
	if err != nil {
		return 0, err
	}
	defer fSrc.Close()

	// Get the size of src file.
	fi, err := fSrc.Stat()
	if err != nil {
		return 0, err
	}

	// Check if src's a regular file.
	if !fi.Mode().IsRegular() {
		return 0, ErrNotFSRegularFile
	}

	// Get total size of src.
	size := fi.Size()

	// Make dest file's dir if it does not exist.
	dir := filepath.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return 0, err
	}

	// Create the dst file.
	fDst, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer fDst.Close()

	// Check if callers need to report progress during IO copy.
	if len(options) > 0 {
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// Number of bytes copied previously.
			// fs.File interface does not require Seek().
			// Resume coping file from file system is not supported.
			0,
			// Options: OnWrittenFunc, Interval.
			options...,
		)

		// Create a multiple writen and dupllicates writes to p.
		mw := io.MultiWriter(fDst, p)

		// Create a channel.
		// Send an empty struct to it to make progress goroutine exit.
		chExit := make(chan struct{}, 1)
		defer func() {
			chExit <- struct{}{}
		}()

		// Starts a new goroutine to report progress until ctx.Done() and chExit receive an empty struct.
		p.Start(ctx, chExit)
		return iocopy.Copy(ctx, mw, fSrc)
	} else {
		return iocopy.Copy(ctx, fDst, fSrc)
	}
}
