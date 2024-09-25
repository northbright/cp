package cp_test

import (
	"context"
	"embed"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/northbright/cp"
	"github.com/northbright/iocopy/progress"
)

var (
	//go:embed copyfsfile.go copyfsfile_test.go
	embededFiles embed.FS
)

func ExampleCopyFSFile() {
	// Example 1. Copy a file in embeded file system and report progress.
	log.Printf("\n============ CopyFSFile Example 1 Begin ============")
	src := "copyfsfile.go"
	dst := filepath.Join(os.TempDir(), "copyfsfile.go")

	log.Printf("cp.CopyFSFile() starts...\nsrc: %v\ndst: %v", src, dst)
	n, err := cp.CopyFSFile(
		// Context.
		context.Background(),
		// Src file system.
		embededFiles,
		// Src in the file system.
		src,
		// Dst.
		dst,
		// OnWrittenFunc to report progress.
		progress.OnWritten(func(total, prev, current int64, percent float32) {
			log.Printf("%v / %v(%.2f%%) coipied", prev+current, total, percent)
		}),
		// Interval to report progress.
		progress.Interval(time.Millisecond*50),
	)

	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFSFile() error: %v", err)
			return
		}
		log.Printf("cp.CopyFSFile() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFSFile() OK, %v bytes copied", n)
	}

	// Remove the files after test's done.
	os.Remove(dst)

	log.Printf("\n------------ CopyFSFile Example 1 End ------------")

	// Output:
}
