package cp_test

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/northbright/cp"
	"github.com/northbright/iocopy"
)

var (
	//go:embed copyfsfile.go copyfsfile_test.go
	embededFiles embed.FS
)

func ExampleCopyFSFileBufferWithProgress() {
	src := "copyfsfile.go"
	dst := filepath.Join(os.TempDir(), "copyfsfile.go")
	buf := make([]byte, 1024*640)

	log.Printf("cp.CopyFSFileBufferWithProgress() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err := cp.CopyFSFileBufferWithProgress(
		// Context.
		context.Background(),
		// Src file system.
		embededFiles,
		// Src in the file system.
		src,
		// Dst.
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
			log.Printf("cp.CopyFSFileBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("cp.CopyFSFileBufferWithProgress() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFSFileBufferWithProgress() OK, %v bytes copied", n)
	}

	// Remove the files after test's done.
	os.Remove(dst)

	// Output:
}
