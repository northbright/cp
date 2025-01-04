package cp

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

var (
	// ErrNotRegularFile represents the error that src is not a regular file.
	ErrNotRegularFile = errors.New("not a regular file")
)

// CopyFileBufferWithProgress copies file from src to dst and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// copied: number of bytes copied previously.
// It can be used to resume the copy.
// 1. Set copied to 0 when call CopyFileBufferWithProgress for the first time.
// 2. User stops the copy and CopyFileBufferWithProgress returns the number of bytes copied and error.
// 3. Check if err == context.Canceled || err == context.DeadlineExceeded.
// 4. Set copied to the "n" return value of previous CopyFileBufferWithProgress when make next call to resume the copy.
// fn: callback on bytes written.
func CopyFileBufferWithProgress(
	ctx context.Context,
	src string,
	dst string,
	buf []byte,
	copied int64,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
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

	return iocopy.CopyBufferWithProgress(ctx, fDst, fSrc, buf, size, copied, fn)
}

// CopyFile copies file from src to dst and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
func CopyFile(ctx context.Context, src, dst string) (n int64, err error) {
	return CopyFileBufferWithProgress(ctx, src, dst, nil, 0, nil)
}

// CopyFileBuffer is buffered version of [CopyFile].
func CopyFileBuffer(ctx context.Context, src, dst string, buf []byte) (n int64, err error) {
	return CopyFileBufferWithProgress(ctx, src, dst, buf, 0, nil)
}

// CopyFileWithProgress is non-buffered version of [CopyFileBufferWithProgress].
func CopyFileWithProgress(
	ctx context.Context,
	src string,
	dst string,
	copied int64,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyFileBufferWithProgress(ctx, src, dst, nil, copied, fn)
}
