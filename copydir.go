package cp

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/northbright/pathelper"
)

// CopyDir copies files and sub-directories from src to dst recursively.
func CopyDir(ctx context.Context, src, dst string) (copied int64, err error) {
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
			// Ignored for copying dir.
		)
		if err != nil {
			return err
		}

		copied += n
		return nil
	})
	return copied, err
}
