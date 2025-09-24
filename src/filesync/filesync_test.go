package filesync

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to create test file with given content and modtime
// helper to create test file with given content and modtime
func writeTestFile(tb testing.TB, path, content string, modtime time.Time) {
	tb.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		tb.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		tb.Fatalf("write failed: %v", err)
	}
	if !modtime.IsZero() {
		_ = os.Chtimes(path, modtime, modtime)
	}
}

func TestFileSync_CopyNewFiles(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")

	writeTestFile(t, filepath.Join(src, "a.txt"), "hello", time.Now())

	fs := NewFileSync(src, dst, false)
	if err := fs.SyncDirs(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("expected hello, got %s", string(data))
	}
}

func TestFileSync_UpdateChangedFiles(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")

	// initial copy
	oldTime := time.Now().Add(-time.Hour)
	writeTestFile(t, filepath.Join(src, "a.txt"), "old", oldTime)
	fs := NewFileSync(src, dst, false)
	_ = fs.SyncDirs()

	// update source with new content
	newTime := time.Now()
	writeTestFile(t, filepath.Join(src, "a.txt"), "new", newTime)
	if err := fs.SyncDirs(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "new" {
		t.Errorf("expected updated file, got %s", data)
	}
}

func TestFileSync_DeleteMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")

	writeTestFile(t, filepath.Join(src, "keep.txt"), "keep", time.Now())
	writeTestFile(t, filepath.Join(dst, "remove.txt"), "remove", time.Now())

	fs := NewFileSync(src, dst, true)
	if err := fs.SyncDirs(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dst, "remove.txt")); !os.IsNotExist(err) {
		t.Error("expected remove.txt to be deleted")
	}
	if _, err := os.Stat(filepath.Join(dst, "keep.txt")); err != nil {
		t.Error("expected keep.txt to remain")
	}
}

func TestFileSync_InvalidPath(t *testing.T) {
	fs := NewFileSync("nonexistent", t.TempDir(), false)
	if err := fs.SyncDirs(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func BenchmarkFileSync_10Files(b *testing.B) {
	benchmarkFileSync(b, 10)
}

func BenchmarkFileSync_100Files(b *testing.B) {
	benchmarkFileSync(b, 100)
}

func BenchmarkFileSync_1000Files(b *testing.B) {
	benchmarkFileSync(b, 1000)
}

func benchmarkFileSync(b *testing.B, n int) {
	tmp := b.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")

	// prepare files
	for i := 0; i < n; i++ {
		// safe, unique file names: file0.txt, file1.txt, ...
		name := filepath.Join(src, "f", fmt.Sprintf("file%d.txt", i))
		writeTestFile(b, name, "content", time.Now())
	}

	fs := NewFileSync(src, dst, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := fs.SyncDirs(); err != nil {
			b.Fatal(err)
		}
	}
}
