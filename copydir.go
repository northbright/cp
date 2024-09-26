package cp

import (
	"context"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/northbright/iocopy/progress"
	"github.com/northbright/pathelper"
)

var (
	// Default interval of OnCopyDirFunc.
	DefaultOnCopyDirInterval = time.Millisecond * 500
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

// OnCopyDirFunc is the callback func to report copy dir progress.
// fileCount: total count of the files in the dir and sub dirs.
// copiedFileCount: count of the copied files.
// totalSize: total size of the files in the dir and sub dirs.
// copiedSize: size of copied files.
// totalPercent: percent of copied size.
// currentFile: name of currently coping file.
// totalOfCurrentFile: size of currently coping file.
// currentOfCurrentFile: copied size of currently coping file.
// percent: percent of copied size of currently coping file.
type OnCopyDirFunc func(
	fileCount,
	copiedFileCount,
	totalSize,
	copiedSize int64,
	totalPercent float32,
	currentFile string,
	totalOfCurrentFile,
	currentOfCurrentFile int64,
	percent float32,
)

type dirCopier struct {
	fn       OnCopyDirFunc
	interval time.Duration
}

// CopyDirOption sets optional parameter to report copy dir progress.
type CopyDirOption func(dc *dirCopier)

// OnCopyDir returns the option to set callback to report copy dir progress.
func OnCopyDir(fn OnCopyDirFunc) CopyDirOption {
	return func(dc *dirCopier) {
		dc.fn = fn
	}
}

// OnDirCopyInterval returns the option to set interval of the callback.
func OnDirCopyInterval(d time.Duration) CopyDirOption {
	return func(dc *dirCopier) {
		dc.interval = d
	}
}

// CopyDir copies files and sub-directories from src to dst recursively.
// ctx: [context.Context].
// src: source dir.
// dst: destination dir.
// options: [CopyDirOption] to report copy dir progress.
func CopyDir(ctx context.Context, src, dst string, options ...CopyDirOption) (copied int64, err error) {
	di, err := DirInfo(src)
	if err != nil {
		return 0, err
	}

	dc := &dirCopier{}
	for _, option := range options {
		option(dc)
	}

	if dc.interval <= 0 {
		dc.interval = DefaultOnCopyDirInterval
	}

	fileCount := di.FileCount
	copiedFileCount := int64(0)
	totalSize := di.TotalSize
	copied = 0

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
		// Make dst file name.
		dstFile := pathelper.ReplacePrefix(path, src, dst)

		if dc.fn != nil {
			// Copy file and report progress.
			n, err := CopyFile(
				// Context.
				ctx,
				// Src.
				path,
				// Dst.
				dstFile,
				// Number of bytes copied previously.
				0,
				// progress.Option to report progress.
				progress.OnWritten(func(total, prev, current int64, percent float32) {
					// Call OnCopyDir callback.
					dc.fn(
						fileCount,
						copiedFileCount,
						totalSize,
						copied,
						progress.Percent(totalSize, 0, copied),
						path,
						total,
						current,
						percent,
					)
				}),
				// Interval to report the progress.
				progress.Interval(dc.interval),
			)
			if err != nil {
				return err
			}

			copied += n
			copiedFileCount += 1

			// Call OnCopyDir callback when a file copied.
			dc.fn(
				fileCount,
				copiedFileCount,
				totalSize,
				copied,
				progress.Percent(totalSize, 0, copied),
				path,
				n,
				n,
				100,
			)

			return nil
		} else {
			// Copy file without reporting progress.
			n, err := CopyFile(ctx, path, dstFile, 0)
			if err != nil {
				return err
			}

			copied += n
			copiedFileCount += 1

			return nil
		}
	})
	return copied, err
}
