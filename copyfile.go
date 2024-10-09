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

	// Default interval of OnCopyFile.
	DefaultOnCopyFileInterval = time.Millisecond * 500
)

type fileCopier struct {
	fn       OnCopyFileFunc
	interval time.Duration
}

// CopyFileOption sets optional parameter to report copy dir progress.
type CopyFileOption func(fc *fileCopier)

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
// copied: number of bytes copied previously. It's used to resume previous copy.
// buf: buffer used for the copy.
// options: [CopyFileOption] used to report progress.
func CopyFileBuffer(ctx context.Context, src, dst string, copied int64, buf []byte, options ...CopyFileOption) (written int64, err error) {
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

	// Set options.
	fc := &fileCopier{}
	for _, option := range options {
		option(fc)
	}

	var writer io.Writer = fDst

	// Check if callers need to report progress during IO copy.
	if fc.fn != nil {
		// Create a progress.
		p := progress.New(
			// Total size.
			size,
			// Number of bytes copied previously.
			copied,
			// OnWrittenFunc option.
			progress.OnWritten(progress.OnWrittenFunc(fc.fn)),
			// Interval option.
			progress.Interval(fc.interval),
		)

		// Create a multiple writen and dupllicates writes to p.
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

	if buf != nil && len(buf) != 0 {
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
// copied: number of bytes copied previously. It's used to resume previous copy.
// options: [CopyFileOption] used to report progress.
func CopyFile(ctx context.Context, src, dst string, copied int64, options ...CopyFileOption) (written int64, err error) {
	return CopyFileBuffer(ctx, src, dst, copied, nil, options...)
}
