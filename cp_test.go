package cp_test

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/northbright/cp"
)

func ExampleCopyFile() {
	ctx := context.Background()

	src := "cp.go"

	dstDir := filepath.Join(os.TempDir(), "cp")
	dst := filepath.Join(dstDir, "cp.go")
	bufSize := uint(16 * 1024 * 1024)

	n, err := cp.CopyFile(ctx, dst, src, bufSize)
	if err != nil {
		log.Printf("CopyFile() error: %v", err)
		return
	}

	log.Printf("CopyFile() OK, %d bytes copied successfully", n)

	// Remove temp file and dir.
	if err := os.RemoveAll(dstDir); err != nil {
		log.Printf("RemoveAll() error: %v", err)
		return
	}

	// Output:
}
