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

// DirInfoData contains dir information.
type DirInfoData struct {
	// Desired file extensions.
	Exts        []string
	FileCount   int64
	SubDirCount int64
	TotalSize   int64
}

// DirInfo returns the dir info.
// dir: directory to get info.
// exts: desired file extensions. Leave it nil or empty for all files.
func DirInfo(dir string, exts []string) (*DirInfoData, error) {
	di := &DirInfoData{}

	for _, ext := range exts {
		di.Exts = append(di.Exts, strings.ToLower(ext))
	}

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

// CopyDirBufferWithProgress copies files and sub-directories from src to dst recursively and returns the number of bytes of copied.
// It accepts [context.Context] to make copy cancalable.
// It also accepts callback function on bytes written to report progress.
// ctx: context to stop the copy.
// src: source dir.
// dst: destination dir.
// exts: desired file extensions. Leave it nil or empty for all files.
// fn: callback on bytes written.
func CopyDirBufferWithProgress(
	ctx context.Context,
	src string,
	dst string,
	exts []string,
	buf []byte,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	di, err := DirInfo(src, exts)
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
		// Check if file ext matched desired exts.
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
// ctx: context to stop the copy.
// src: source dir.
// dst: destination dir.
// exts: desired file extensions. Leave it nil or empty for all files.
func CopyDir(ctx context.Context, src, dst string, exts []string) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, exts, nil, nil)
}

// CopyDirBuffer is buffered version of [CopyDir].
func CopyDirBuffer(ctx context.Context, src, dst string, exts []string, buf []byte) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, exts, buf, nil)
}

// CopyDirWithProgress is non-buffered version of [CopyDirBufferWithProgress].
func CopyDirWithProgress(
	ctx context.Context,
	src string,
	dst string,
	exts []string,
	fn iocopy.OnWrittenFunc) (n int64, err error) {
	return CopyDirBufferWithProgress(ctx, src, dst, exts, nil, fn)
}
