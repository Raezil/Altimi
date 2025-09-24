package filesync

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type FileSync struct {
	source        string
	target        string
	deleteMissing bool
}

func NewFileSync(source, target string, deleteMissing bool) *FileSync {
	return &FileSync{
		source:        source,
		target:        target,
		deleteMissing: deleteMissing,
	}
}

func (fs *FileSync) SyncDirs() error {
	// Walk through files in source
	err := filepath.WalkDir(fs.source, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error accessing %s: %v", path, err)
			return nil // continue despite the error
		}

		relPath, _ := filepath.Rel(fs.source, path)
		targetPath := filepath.Join(fs.target, relPath)

		// Handle directories
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

		if tgtInfo, err := os.Stat(targetPath); os.IsNotExist(err) {
			copy = true // file does not exist in target
		} else if err == nil {
			if !fs.sameFile(srcInfo, tgtInfo) {
				copy = true // differs (time or size)
			}
		} else {
			log.Printf("❌ Problem reading %s: %v", targetPath, err)
		}

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

	// Optionally delete missing files
	if fs.deleteMissing {
		err = filepath.WalkDir(fs.target, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				log.Printf("Error accessing %s: %v", path, err)
				return nil
			}
			relPath, _ := filepath.Rel(fs.target, path)
			srcPath := filepath.Join(fs.source, relPath)

			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				if d.IsDir() {
					// remove only empty directories
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

func (fs *FileSync) sameFile(src, tgt os.FileInfo) bool {
	return src.Size() == tgt.Size() && src.ModTime().Equal(tgt.ModTime())
}

func (fs *FileSync) copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	// preserve modification time
	if srcInfo, err := os.Stat(src); err == nil {
		os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	}

	return nil
}
