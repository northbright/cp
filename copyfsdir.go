package cp

import (
	"context"
	"io/fs"

	"github.com/northbright/iocopy/progress"
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

// CopyFSDir copies files and sub-directories of src from the file system to dst recursively.
// ctx: [context.Context].
// fsys: file system of src.
// src: source dir.
// dst: destination dir.
// options: [CopyDirOption] to report copy dir progress.
func CopyFSDir(ctx context.Context, fsys fs.FS, src, dst string, options ...CopyDirOption) (copied int64, err error) {
	di, err := FSDirInfo(fsys, src)
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
		// Make dst file name.
		dstFile := pathelper.ReplacePrefix(path, src, dst)

		if dc.fn != nil {
			// Copy file and report progress.
			n, err := CopyFSFile(
				// Context.
				ctx,
				// File system.
				fsys,
				// Src.
				path,
				// Dst.
				dstFile,
				// progress.Option to report progress.
				progress.OnWritten(func(total, prev, current int64, percent float32) {
					// Call OnCopyFSDir callback.
					dc.fn(
						fileCount,
						copiedFileCount,
						totalSize,
						// Use copied + current as new copied.
						copied+current,
						progress.Percent(totalSize, 0, copied+current),
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

			// Call OnCopyFSDir callback when a file copied.
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
			n, err := CopyFSFile(ctx, fsys, path, dstFile)
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
