package filesync

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// FileSync represents a one-way synchronization job
// from a source directory to a target directory.
// If deleteMissing is true, extra files in the target
// (not present in source) will be removed.
type FileSync struct {
	source        string
	target        string
	deleteMissing bool
}

// NewFileSync constructs a FileSync instance.
//
// Parameters:
//   - source: directory path to copy files from
//   - target: directory path to copy files into
//   - deleteMissing: whether to remove files from target
//     if they don’t exist in source
func NewFileSync(source, target string, deleteMissing bool) *FileSync {
	return &FileSync{
		source:        source,
		target:        target,
		deleteMissing: deleteMissing,
	}
}

// SyncDirs synchronizes the contents of source → target.
//
// Behavior:
//  1. Walks the source directory.
//  2. Creates missing directories in target.
//  3. Copies new or updated files into target.
//  4. Optionally deletes files/dirs in target
//     that do not exist in source (if deleteMissing is set).
//
// Returns an error only if the initial directory walk fails
// or if target cleanup encounters issues; per-file errors
// are logged but do not stop the process.
func (fs *FileSync) SyncDirs() error {
	// Walk through all entries in source
	err := filepath.WalkDir(fs.source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip problem entries but continue walking
			log.Printf("Error accessing %s: %v", path, err)
			return nil
		}

		// Build target path relative to source root
		relPath, _ := filepath.Rel(fs.source, path)
		targetPath := filepath.Join(fs.target, relPath)

		// Handle directories: ensure existence in target
		if d.IsDir() {
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				if mkErr := os.MkdirAll(targetPath, 0755); mkErr != nil {
					log.Printf("❌ Failed to create directory %s: %v", targetPath, mkErr)
				} else {
					log.Printf("📂 Created directory: %s", targetPath)
				}
			}
			return nil
		}

		// Handle files
		copy := false
		srcInfo, err := os.Stat(path)
		if err != nil {
			log.Printf("❌ Could not read file info for %s: %v", path, err)
			return nil
		}

		// Determine whether to copy:
		// - Missing in target
		// - Different size or modification time
		if tgtInfo, err := os.Stat(targetPath); os.IsNotExist(err) {
			copy = true
		} else if err == nil {
			if !fs.sameFile(srcInfo, tgtInfo) {
				copy = true
			}
		} else {
			log.Printf("❌ Problem reading %s: %v", targetPath, err)
		}

		// Perform copy if flagged
		if copy {
			if err := fs.copyFile(path, targetPath); err != nil {
				log.Printf("❌ Error copying %s → %s: %v", path, targetPath, err)
			} else {
				log.Printf("📄 Copied/Updated: %s → %s", path, targetPath)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Optionally clean up extra files in target
	if fs.deleteMissing {
		err = filepath.WalkDir(fs.target, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				log.Printf("Error accessing %s: %v", path, err)
				return nil
			}

			// Find matching path in source
			relPath, _ := filepath.Rel(fs.target, path)
			srcPath := filepath.Join(fs.source, relPath)

			// Remove target entry if it doesn’t exist in source
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				if d.IsDir() {
					// Attempt to remove empty directory
					if rmErr := os.Remove(path); rmErr == nil {
						log.Printf("🗑️ Removed empty directory: %s", path)
					}
				} else {
					if rmErr := os.Remove(path); rmErr == nil {
						log.Printf("🗑️ Removed file: %s", path)
					}
				}
			}
			return nil
		})
	}

	return err
}

// sameFile compares two files by size and modification time.
// Returns true if they appear identical.
func (fs *FileSync) sameFile(src, tgt os.FileInfo) bool {
	return src.Size() == tgt.Size() && src.ModTime().Equal(tgt.ModTime())
}

// copyFile copies src → dst, creating parent directories if needed.
// The modification time of the source file is preserved on the target.
func (fs *FileSync) copyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Open source file
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Create or truncate target file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy contents
	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// Preserve modification time from source
	if srcInfo, err := os.Stat(src); err == nil {
		os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}

	return nil
}
