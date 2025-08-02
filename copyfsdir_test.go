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
	//go:embed assets
	assets embed.FS
)

func ExampleCopyFSDirBufferWithProgress() {
	src := "assets"
	dst := filepath.Join(os.TempDir(), "copied_assets")
	buf := make([]byte, 1024*640)

	log.Printf("cp.CopyFSDirBufferWithProgress() starts...\nsrcFS: %v\nsrc: %v\ndst: %v", "embed.FS for assets", src, dst)

	n, err := cp.CopyFSDirBufferWithProgress(
		// Context.
		context.Background(),
		// File system.
		assets,
		// Src.
		src,
		// Dst.
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
		log.Printf("cp.CopyFSDirBufferWithProgress() error: %v", err)
		return
	}
	if err != nil {
		if err != context.Canceled && err != context.DeadlineExceeded {
			log.Printf("cp.CopyFSDirBufferWithProgress() error: %v", err)
			return
		}
		log.Printf("cp.CopyFSDirBufferWithProgress() stopped, cause: %v. %v bytes copied", err, n)
	} else {
		log.Printf("cp.CopyFSDirBufferWithProgress() OK, %v bytes copied", n)
	}

	// Remove dir after test.
	os.RemoveAll(dst)

	// Output:
}
