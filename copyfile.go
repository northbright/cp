package cp

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/iocopy/progress"
	"github.com/northbright/pathelper"
)

var (
	// ErrNotRegularFile represents the error that src is not a regular file.
	ErrNotRegularFile = errors.New("not regular file")
)

// CopyFile copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// src: source file.
// dst: destination file.
// copied: number of bytes copied previously. It's used to resume previous copy.
// options: [progress.Option] used to report progress.
func CopyFile(ctx context.Context, src, dst string, copied int64, options ...progress.Option) (written int64, err error) {
	// Get src file info.
	fi, err := os.Lstat(src)
	if err != nil {
		return 0, err
	}

	// Check if src's a regular file.
	if !fi.Mode().IsRegular() {
		return 0, ErrNotRegularFile
	}

	// Get the source file's size.
	size := fi.Size()

	// Make dest file's dir if it does not exist.
	dir := filepath.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return 0, err
	}

	fSrc, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer fSrc.Close()

	var fDst *os.File

	if copied > 0 {
		if fDst, err = os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return 0, err
		}
		defer fDst.Close()

		if _, err = fSrc.Seek(copied, 0); err != nil {
			return 0, err
		}

		if _, err = fDst.Seek(copied, 0); err != nil {
			return 0, err
		}
	} else {
		if copied < 0 {
			copied = 0
		}

		if fDst, err = os.Create(dst); err != nil {
			return 0, err
		}
		defer fDst.Close()
	}

	// Check if callers need to report progress during IO copy.
	if len(options) > 0 {
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// Number of bytes copied previously.
			copied,
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
