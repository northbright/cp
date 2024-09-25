package cp_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/cp"
	"github.com/northbright/download"
	"github.com/northbright/iocopy/progress"
)

func ExampleCopyFile() {
	// Example 1. Copy a file and report progress.
	log.Printf("\n============ CopyFile Example 1 Begin ============")

	// Download a file.
	url := "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.23.1.darwin-amd64.pkg")

	log.Printf("download.Download() starts...\nurl: %v\ndst: %v", url, dst)

	n, err := download.Download(
		// Context.
		context.Background(),
		// URL to download.
		url,
		// Destination.
		dst,
		// Number of bytes copied previously.
		0,
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

	// Copy the downloaded file to another file.
	src := dst
	dst = filepath.Join(os.TempDir(), "go.pkg")

	log.Printf("cp.CopyFile() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err = cp.CopyFile(
		// Context.
		context.Background(),
		// Source file.
		src,
		// Destination file.
		dst,
		// Number of bytes copied previously.
		0,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
		// Interval to report progress.
		progress.Interval(time.Millisecond*50),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFile() error: %v", err)
			return
		}
		log.Printf("cp.CopyFile() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFile() OK, %v bytes copied", n)
	}

	log.Printf("\n------------ CopyFile Example 1 End ------------")

	// Example 2. Stop a copy and resume it.
	log.Printf("\n============ CopyFile Example 2 Begin ============")

	// Create a timeout to emulate user's cancelation.
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cancel()

	log.Printf("cp.CopyFile() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err = cp.CopyFile(
		// Context.
		ctx,
		// Source file.
		src,
		// Destination file.
		dst,
		// Number of bytes copied previously.
		0,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
		// Interval to report progress.
		progress.Interval(time.Millisecond*50),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFile() error: %v", err)
			return
		}
		log.Printf("cp.CopyFile() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFile() OK, %v bytes copied", n)
	}

	log.Printf("cp.CopyFile() starts again to resume coping...\nsrc: %v\ndst: %v\ncopied: %v", src, dst, n)

	// Set copied to n to resume the copy.
	n2, err := cp.CopyFile(
		// Context.
		context.Background(),
		// Source file.
		src,
		// Destination file.
		dst,
		// Number of bytes copied previously.
		n,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
		// Interval to report progress.
		progress.Interval(time.Millisecond*50),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFile() error: %v", err)
			return
		}
		log.Printf("cp.CopyFile() stopped, cause: %v. %v bytes copied", err, n2)
	} else {
		log.Printf("cp.CopyFile() OK, %v bytes copied", n2)
	}

	log.Printf("total %v bytes copied", n+n2)

	// Remove the files after test's done.
	os.Remove(dst)
	os.Remove(src)

	log.Printf("\n------------ CopyFile Example 2 End ------------")

	// Output:
}
