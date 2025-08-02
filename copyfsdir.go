package cp

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

// FSDirInfo returns the dir info.
func FSDirInfo(fsys fs.FS, dir string, exts []string) (*DirInfoData, error) {
	di := &DirInfoData{}

	for _, ext := range exts {
		di.Exts = append(di.Exts, strings.ToLower(ext))
	}

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

		if len(di.Exts) > 0 {
			for _, ext := range di.Exts {
				if strings.ToLower(filepath.Ext(d.Name())) == ext {
					di.FileCount += 1
					di.TotalSize += fi.Size()
					return nil
				}
			}
			return nil
		} else {
			di.FileCount += 1
			di.TotalSize += fi.Size()
			return nil
		}
	})

	return di, err
}

// CopyFSDirBufferWithProgress copies files and sub-directories of src from the file system to dst recursively and returns the number of bytes copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// ctx: context to stop the copy.
// fsys: file system.
// src: source dir.
// dst: destination dir.
// exts: desired file extensions. Leave it nil or empty for all files.
// fn: callback on bytes written.
func CopyFSDirBufferWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	exts []string,
	buf []byte,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	di, err := FSDirInfo(fsys, src, exts)
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
		// Check if file ext matches desired exts.
		matched := false
		if len(di.Exts) > 0 {
			for _, ext := range di.Exts {
				if strings.ToLower(filepath.Ext(d.Name())) == ext {
					matched = true
					break
				}
			}
		} else {
			matched = true
		}

		// Skip if ext is not matched.
		if !matched {
			return nil
		}

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
// ctx: context to stop the copy.
// fsys: file system.
// src: source dir.
// dst: destination dir.
// exts: desired file extensions. Leave it nil or empty for all files.
func CopyFSDir(ctx context.Context, fsys fs.FS, src, dst string, exts []string) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, exts, nil, nil)
}

// CopyFSDirBuffer is buffered version of [CopyFSDir].
func CopyFSDirBuffer(ctx context.Context, fsys fs.FS, src, dst string, exts []string, buf []byte) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, exts, buf, nil)
}

// CopyFSDirWithProgress is non-buffered version of [CopyFSDirBufferWithProgress].
func CopyFSDirWithProgress(
	ctx context.Context,
	fsys fs.FS,
	src string,
	dst string,
	exts []string,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyFSDirBufferWithProgress(ctx, fsys, src, dst, exts, nil, fn)
}
