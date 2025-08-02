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
	dir := filepath.Join(os.TempDir(), "go-releases")
	os.MkdirAll(dir, 0755)

	type GoRelease struct {
		os  string
		ver string
		url string
	}

	releases := []GoRelease{
		GoRelease{
			os:  "macOS",
			ver: "1.23.1",
			url: "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg",
		},
		GoRelease{
			os:  "linux",
			ver: "1.23.1",
			url: "https://golang.google.cn/dl/go1.23.1.linux-amd64.tar.gz",
		},
		GoRelease{
			os:  "windows",
			ver: "1.23.1",
			url: "https://golang.google.cn/dl/go1.23.1.windows-amd64.msi",
		},
		GoRelease{
			os:  "macOS",
			ver: "1.24.5",
			url: "https://golang.google.cn/dl/go1.24.5.darwin-amd64.pkg",
		},
		GoRelease{
			os:  "linux",
			ver: "1.24.5",
			url: "https://golang.google.cn/dl/go1.24.5.linux-amd64.tar.gz",
		},
		GoRelease{
			os:  "windows",
			ver: "1.24.5",
			url: "https://golang.google.cn/dl/go1.24.5.windows-amd64.msi",
		},
	}

	// Download files to src dir.
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

	// Copy dir for all files.
	log.Printf("cp.CopyDirBufferWithProgress() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err := cp.CopyDirBufferWithProgress(
		// Context.
		context.Background(),
		// Source dir.
		src,
		// Destination dir.
		dst,
		// Desired file extensions. Leave it nil or empty for all files.
		nil,
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

	// Remove dst dir after test.
	os.RemoveAll(dst)

	// Copy dir for desired file extensions.
	dst = filepath.Join(os.TempDir(), "go-gz-pkg")
	exts := []string{".gz", ".pkg"}
	log.Printf("cp.CopyDirBufferWithProgress() for \".gz\" and \".pkg\" starts...\nsrc: %v\ndst: %v", src, dst)
	n, err = cp.CopyDirBufferWithProgress(
		// Context.
		context.Background(),
		// Source dir.
		src,
		// Destination dir.
		dst,
		// Desired file extensions. Leave it nil or empty for all files.
		exts,
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

	// Remove dirs after all tests.
	os.RemoveAll(src)
	os.RemoveAll(dst)

	// Output:
}
