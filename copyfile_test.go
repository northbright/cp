package cp_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/cp"
	"github.com/northbright/iocopy"
	"github.com/northbright/pathelper"
)

func download(url, dst string) (int64, error) {
	// Create parent dir of dst if it does not exist.
	dir := filepath.Dir(dst)
	if err := pathelper.CreateDirIfNotExists(dir, 0755); err != nil {
		return 0, err
	}

	f, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Use io.Copy to download the remote file.
	return io.Copy(f, resp.Body)
}

func ExampleCopyFileBufferWithProgress() {
	// Download a file as source file.
	url := "https://golang.google.cn/dl/go1.23.1.darwin-amd64.pkg"
	dst := filepath.Join(os.TempDir(), "go1.23.1.darwin-amd64.pkg")

	n, err := download(url, dst)
	if err != nil {
		log.Printf("download() error: %v", err)
		return
	}

	// Copy the downloaded file to another file.
	src := dst
	dst = filepath.Join(os.TempDir(), "go.pkg")

	buf := make([]byte, 1024*640)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*80)
	defer cancel()

	log.Printf("cp.CopyFileBufferWithProgress() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err = cp.CopyFileBufferWithProgress(
		// Context.
		ctx,
		// Source file.
		src,
		// Destination file.
		dst,
		// Buffer.
		buf,
		// Number of bytes copied previously.
		0,
		// Callback to report progress.
		iocopy.OnWrittenFunc(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFileBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("cp.CopyFileBufferWithProgress() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFileBufferWithProgress() OK, %v bytes copied", n)
		return
	}

	log.Printf("cp.CopyFileBufferWithProgress() starts again to resume coping...\nsrc: %v\ndst: %v\ncopied: %v", src, dst, n)

	// Set copied to n to resume the copy.
	n2, err := cp.CopyFileBufferWithProgress(
		// Context.
		context.Background(),
		// Source file.
		src,
		// Destination file.
		dst,
		// Buffer.
		buf,
		// Number of bytes copied previously.
		n,
		// Callback to report progress.
		iocopy.OnWrittenFunc(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFileBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("cp.CopyFileBufferWithProgress() stopped, cause: %v. %v bytes copied", err, n2)
	} else {
		log.Printf("cp.CopyFileBufferWithProgress() OK, %v bytes copied", n2)
	}

	log.Printf("total %v bytes copied", n+n2)

	// Remove the files after test's done.
	os.Remove(dst)
	os.Remove(src)

	// Output:
}
