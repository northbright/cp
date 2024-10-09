package cp_test

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/cp"
)

var (
	//go:embed assets
	assets embed.FS
)

func ExampleCopyFSDir() {
	// Example 1. Copy dir from embeded file system to dst.
	log.Printf("\n============ CopyFSDir Example 1 Begin ============")

	src := "assets"
	dst := filepath.Join(os.TempDir(), "copied_assets")

	log.Printf("cp.CopyFSDir() starts...\nsrcFS: %v\nsrc: %v\ndst: %v", "embed.FS for assets", src, dst)

	n, err := cp.CopyFSDir(
		// Context.
		context.Background(),
		// File system.
		assets,
		// Src.
		src,
		// Dst.
		dst,
	)
	if err != nil {
		log.Printf("cp.CopyFSDir() error: %v", err)
		return
	}
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFSDir() error: %v", err)
			return
		}
		log.Printf("cp.CopyFSDir() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFSDir() OK, %v bytes copied", n)
	}

	log.Printf("\n------------ CopyFSDir Example 1 End ------------")

	// Example 2. Copy dir from embeded file system to dst and report progress.
	log.Printf("\n============ CopyFSDir Example 2 Begin ============")

	log.Printf("cp.CopyFSDir() starts...\nsrcFS: %v\nsrc: %v\ndst: %v", "embed.FS for assets", src, dst)
	n, err = cp.CopyFSDir(
		// Context.
		context.Background(),
		// File system.
		assets,
		// Src.
		src,
		// Dst.
		dst,
		// CopyDirOption to report progress.
		cp.OnCopyDir(func(
			fileCount,
			copiedFileCount,
			totalSize,
			copiedSize int64,
			totalPercent float32,
			currentFile string,
			totalOfCurrentFile,
			currentOfCurrentFile int64,
			percent float32,
		) {
			log.Printf("\n******************\n%v / %v files copied\n%v / %v(%.2f%%) bytes copied\ncurrent coping file: %v\n%v / %v(%.2f%%) bytes copied",
				copiedFileCount,
				fileCount,
				copiedSize,
				totalSize,
				totalPercent,
				currentFile,
				currentOfCurrentFile,
				totalOfCurrentFile,
				percent,
			)
		}),
		// Interval to report progress.
		cp.OnCopyDirInterval(time.Millisecond*10),
	)

	if err != nil {
		log.Printf("cp.CopyFSDir() error: %v", err)
		return
	}
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFSDir() error: %v", err)
			return
		}
		log.Printf("cp.CopyFSDir() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFSDir() OK, %v bytes copied", n)
	}

	log.Printf("\n------------ CopyFSDir Example 2 End ------------")

	// Remove dir after test.
	os.RemoveAll(dst)

	// Output:
}
