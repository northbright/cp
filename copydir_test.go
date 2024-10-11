package cp_test

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/northbright/cp"
	"github.com/northbright/download"
)

func ExampleCopyDir() {
	// Example 1. Download remote files to a temp dir and copy the dir to another one.
	log.Printf("\n============ CopyDir Example 1 Begin ============")

	dir := filepath.Join(os.TempDir(), "go1.23.1")
	os.MkdirAll(dir, 0755)

	type GoRelease struct {
		os  string
		url string
	}

	releases := []GoRelease{
		GoRelease{
			os:  "macOS",
			url: "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg",
		},
		GoRelease{
			os:  "linux",
			url: "https://golang.google.cn/dl/go1.23.1.linux-amd64.tar.gz",
		},
		GoRelease{
			os:  "windows",
			url: "https://golang.google.cn/dl/go1.23.1.windows-amd64.msi",
		},
	}

	for _, release := range releases {
		dst := filepath.Join(dir, release.os, path.Base(release.url))

		log.Printf("download.Download() starts...\nurl: %v\ndst: %v", release.url, dst)
		n, err := download.Download(
			// Context.
			context.Background(),
			// URL to download.
			release.url,
			// Destination.
			dst,
		)

		if err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				log.Printf("download.Download() error: %v", err)
				return
			}
			log.Printf("download.Download() stopped, cause: %v. %v bytes downloaded", err, n)
		} else {
			log.Printf("download.Download() OK, %v bytes downloaded", n)
		}
	}

	src := dir
	dst := filepath.Join(os.TempDir(), "go")

	log.Printf("cp.CopyDir() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err := cp.CopyDir(
		// Context.
		context.Background(),
		// Source dir.
		src,
		// Destination dir.
		dst,
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyDir() error: %v", err)
			return
		}
		log.Printf("cp.CopyDir() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyDir() OK, %v bytes copied", n)
	}

	log.Printf("\n------------ CopyDir Example 1 End ------------")

	// Example 2. Copy dir and report progress.
	log.Printf("\n============ CopyDir Example 2 Begin ============")

	log.Printf("cp.CopyDir() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err = cp.CopyDir(
		// Context.
		context.Background(),
		// Source dir.
		src,
		// Destination dir.
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
		cp.OnCopyDirInterval(time.Millisecond*50),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyDir() error: %v", err)
			return
		}
		log.Printf("cp.CopyDir() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyDir() OK, %v bytes copied", n)
	}

	log.Printf("\n------------ CopyDir Example 2 End ------------")

	// Remove dirs after test.
	os.RemoveAll(src)
	os.RemoveAll(dst)

	// Output:
}
