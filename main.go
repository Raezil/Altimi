package main

import (
	"filesync"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	deleteMissing bool
)

func main() {
	// CLI flags
	flag.BoolVar(&deleteMissing, "delete-missing", false, "Delete files from target that do not exist in source")
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalf("Usage: %s [--delete-missing] <source_dir> <target_dir>", os.Args[0])
	}

	sourceDir := flag.Arg(0)
	targetDir := flag.Arg(1)

	// Check if directories exist
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Fatalf("Source directory does not exist: %s", sourceDir)
	}
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		log.Fatalf("Target directory does not exist: %s", targetDir)
	}

	fs := filesync.NewFileSync(sourceDir, targetDir, deleteMissing)

	// Synchronization
	if err := fs.SyncDirs(); err != nil {
		log.Fatalf("Error during synchronization: %v", err)
	}

	fmt.Println("✅ Synchronization completed successfully.")
}
