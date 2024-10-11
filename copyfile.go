package cp

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/iocopy"
	"github.com/northbright/iocopy/progress"
	"github.com/northbright/pathelper"
)

var (
	// ErrNotRegularFile represents the error that src is not a regular file.
	ErrNotRegularFile = errors.New("not a regular file")
)

type fileCopier struct {
	copied   int64
	fn       OnCopyFileFunc
	interval time.Duration
}

// CopyFileOption sets optional parameter to report copy dir progress.
type CopyFileOption func(fc *fileCopier)

// Copied returns an option to set the number of bytes copied previously.
// It's used to resume previous copy.
func Copied(copied int64) CopyFileOption {
	return func(fc *fileCopier) {
		fc.copied = copied
	}
}

// OnCopyFileFunc is the callback function when bytes are copied successfully.
// See [progress.OnWrittenFunc].
type OnCopyFileFunc progress.OnWrittenFunc

// OnCopyFile returns the option to set callback to report progress.
func OnCopyFile(fn OnCopyFileFunc) CopyFileOption {
	return func(fc *fileCopier) {
		fc.fn = fn
	}
}

// OnCopyFileInterval returns the option to set interval of the callback.
// If no interval option specified, it'll use [DefaultOnCopyFileInterval].
func OnCopyFileInterval(d time.Duration) CopyFileOption {
	return func(fc *fileCopier) {
		fc.interval = d
	}
}

// CopyFileBuffer copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// src: source file.
// dst: destination file.
// buf: buffer used for the copy.
// options: [CopyFileOption] used to report progress.
func CopyFileBuffer(ctx context.Context, src, dst string, buf []byte, options ...CopyFileOption) (written int64, err error) {
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

	// Set options.
	fc := &fileCopier{}
	for _, option := range options {
		option(fc)
	}

	if fc.copied > 0 {
		if fDst, err = os.OpenFile(dst, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return 0, err
		}
		defer fDst.Close()

		if _, err = fSrc.Seek(fc.copied, 0); err != nil {
			return 0, err
		}

		if _, err = fDst.Seek(fc.copied, 0); err != nil {
			return 0, err
		}
	} else {
		if fc.copied < 0 {
			fc.copied = 0
		}

		if fDst, err = os.Create(dst); err != nil {
			return 0, err
		}
		defer fDst.Close()
	}

	var writer io.Writer = fDst

	// Check if callers need to report progress during IO copy.
	if fc.fn != nil {
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// OnWrittenFunc.
			progress.OnWrittenFunc(fc.fn),
			// Option to set number of bytes copied previously.
			progress.Prev(fc.copied),
			// Option to set interval.
			progress.Interval(fc.interval),
		)

		// Create a multiple writer and dupllicates writes to p.
		writer = io.MultiWriter(fDst, p)

		// Create a channel.
		// Send an empty struct to it to make progress goroutine exit.
		chExit := make(chan struct{}, 1)
		defer func() {
			chExit <- struct{}{}
		}()

		// Starts a new goroutine to report progress until ctx.Done() and chExit receive an empty struct.
		p.Start(ctx, chExit)
	}

	if len(buf) != 0 {
		return iocopy.CopyBuffer(ctx, writer, fSrc, buf)
	} else {
		return iocopy.Copy(ctx, writer, fSrc)
	}
}

// CopyFile copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// src: source file.
// dst: destination file.
// options: [CopyFileOption] used to report progress.
func CopyFile(ctx context.Context, src, dst string, options ...CopyFileOption) (written int64, err error) {
	return CopyFileBuffer(ctx, src, dst, nil, options...)
}
