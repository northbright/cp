package cp

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

var (
	// ErrNotFSRegularFile represents the error that src is not a regular file in the file system.
	ErrNotFSRegularFile = errors.New("not a regular file in file system")
)

// CopyFSFileBufferWithProgress copies file from src to dst and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// fn: callback on bytes written.
func CopyFSFileBufferWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	buf []byte,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
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

	return iocopy.CopyBufferWithProgress(ctx, fDst, fSrc, buf, size, 0, fn)
}

// CopyFSFile copies file from src to dst and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
func CopyFSFile(ctx context.Context, fsys fs.FS, src, dst string) (n int64, err error) {
	return CopyFSFileBufferWithProgress(ctx, fsys, src, dst, nil, nil)
}

// CopyFSFileBuffer is buffered version of [CopyFSFile].
func CopyFSFileBuffer(ctx context.Context, fsys fs.FS, src, dst string, buf []byte) (n int64, err error) {
	return CopyFSFileBufferWithProgress(ctx, fsys, src, dst, buf, nil)
}

// CopyFSFileWithProgress is non-buffered version of [CopyFSFileBufferWithProgress].
func CopyFSFileWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyFSFileBufferWithProgress(ctx, fsys, src, dst, nil, fn)
}
