package cp

import (
	"context"
	"io/fs"
	"os"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

// FSDirInfo returns the dir info.
func FSDirInfo(fsys fs.FS, dir string) (*DirInfoData, error) {
	di := &DirInfoData{}

	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
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

// CopyFSDirBufferWithProgress copies files and sub-directories of src from the file system to dst recursively and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// fn: callback on bytes written.
func CopyFSDirBufferWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	buf []byte,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	di, err := FSDirInfo(fsys, src)
	if err != nil {
		return 0, err
	}

	totalSize := di.TotalSize
	copied := int64(0)

	err = fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
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
		fSrc, err := fsys.Open(path)
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

// CopyFSDir copies files and sub-directories of src from the file system to dst recursively and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
func CopyFSDir(ctx context.Context, fsys fs.FS, src, dst string) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, nil, nil)
}

// CopyFSDirBuffer is buffered version of [CopyFSDir].
func CopyFSDirBuffer(ctx context.Context, fsys fs.FS, src, dst string, buf []byte) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, buf, nil)
}

// CopyFSDirWithProgress is non-buffered version of [CopyFSDirBufferWithProgress].
func CopyFSDirWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, nil, fn)
}
