package cp_test

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/northbright/cp"
	"github.com/northbright/iocopy"
)

func ExampleCopyDirBufferWithProgress() {
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
		_, err := download(release.url, dst)
		if err != nil {
			log.Printf("download() error: %v", err)
			return
		}
	}

	src := dir
	dst := filepath.Join(os.TempDir(), "go")
	buf := make([]byte, 1024*640)

	log.Printf("cp.CopyDirBufferWithProgress() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err := cp.CopyDirBufferWithProgress(
		// Context.
		context.Background(),
		// Source dir.
		src,
		// Destination dir.
		dst,
		// Buffer.
		buf,
		// Callback to report progress.
		iocopy.OnWrittenFunc(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyDirBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("cp.CopyDirBufferWithProgress() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyDirBufferWithProgress() OK, %v bytes copied", n)
	}

	// Remove dirs after test.
	os.RemoveAll(src)
	os.RemoveAll(dst)

	// Output:
}
