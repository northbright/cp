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

// CopyFSFileBuffer copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// fsys: file system of src.
// src: source file path in the file system.
// dst: destination file.
// buf: buffer used for the copy.
// options: [CopyFileOption] used to report progress.
func CopyFSFileBuffer(ctx context.Context, fsys fs.FS, src, dst string, buf []byte, options ...CopyFileOption) (written int64, err error) {
	// Open the src file.
	fSrc, err := fsys.Open(src)
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
			// OnWrittenFunc.
			progress.OnWrittenFunc(fc.fn),
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

	if len(buf) != 0 {
		return iocopy.CopyBuffer(ctx, writer, fSrc, buf)
	} else {
		return iocopy.Copy(ctx, writer, fSrc)
	}
}

// CopyFSFile copies file from src to dst.
// It returns the number of bytes copied.
// ctx: [context.Context].
// fsys: file system of src.
// src: source file path in the file system.
// dst: destination file.
// options: [CopyFileOption] used to report progress.
func CopyFSFile(ctx context.Context, fsys fs.FS, src, dst string, options ...CopyFileOption) (written int64, err error) {
	return CopyFSFileBuffer(ctx, fsys, src, dst, nil, options...)
}
