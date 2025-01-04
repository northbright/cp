package cp

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

// DirInfoData contains dir information.
type DirInfoData struct {
	FileCount   int64
	SubDirCount int64
	TotalSize   int64
}

// DirInfo returns the dir info.
func DirInfo(dir string) (*DirInfoData, error) {
	di := &DirInfoData{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		// Check err first.
		// d is nil while the err is "no such file or directory".
		if err != nil {
			return err
		}

		// d is a dir.
		if d.IsDir() {
			di.SubDirCount += 1
			return nil
		}

		// d is a file.
		fi, err := d.Info()
		if err != nil {
			return err
		}
		di.FileCount += 1
		di.TotalSize += fi.Size()
		return nil
	})

	return di, err
}

// CopyDirBufferWithProgress copies files and sub-directories from src to dst recursively and returns the number of bytes of copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// fn: callback on bytes written.
func CopyDirBufferWithProgress(
	ctx context.Context,
	src string,
	dst string,
	buf []byte,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	di, err := DirInfo(src)
	if err != nil {
		return 0, err
	}

	totalSize := di.TotalSize
	copied := int64(0)

	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		// Check err first.
		// d is nil while the err is "no such file or directory".
		if err != nil {
			return err
		}

		// d is a dir.
		if d.IsDir() {
			// Create the dir even if the source dir is empty.
			dstDir := pathelper.ReplacePrefix(path, src, dst)
			return pathelper.CreateDirIfNotExists(dstDir, 0755)
		}

		// d is a file.
		fSrc, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fSrc.Close()

		// Make dst file name.
		dstFile := pathelper.ReplacePrefix(path, src, dst)
		fDst, err := os.Create(dstFile)
		if err != nil {
			return err
		}
		defer fDst.Close()

		n, err := iocopy.CopyBufferWithProgress(
			// Context.
			ctx,
			// Dst.
			fDst,
			// Src.
			fSrc,
			// Buffer.
			buf,
			// Total size of all files in the dir.
			totalSize,
			// Bytes of copied files.
			copied,
			// Callback to report progress.
			fn,
		)
		if err != nil {
			return err
		}
		copied += n
		return nil
	})
	return copied, err
}

// CopyDir copies files and sub-directories from src to dst recursively and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
func CopyDir(ctx context.Context, src, dst string) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, nil, nil)
}

// CopyDirBuffer is buffered version of [CopyDir].
func CopyDirBuffer(ctx context.Context, src, dst string, buf []byte) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, buf, nil)
}

// CopyDirWithProgress is non-buffered version of [CopyDirBufferWithProgress].
func CopyDirWithProgress(
	ctx context.Context,
	src string,
	dst string,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, nil, fn)
}
