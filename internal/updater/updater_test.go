package updater

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateModelCreatesBackup(t *testing.T) {
	parentDir := t.TempDir()
	targetDir := filepath.Join(parentDir, "Rime")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("create target dir: %v", err)
	}
	existingPath := filepath.Join(targetDir, "default.custom.yaml")
	if err := os.WriteFile(existingPath, []byte("old"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := UpdateModel([]byte("model"), targetDir); err != nil {
		t.Fatalf("UpdateModel returned error: %v", err)
	}

	modelPath := filepath.Join(targetDir, "wanxiang-lts-zh-hans.gram")
	if data, err := os.ReadFile(modelPath); err != nil || string(data) != "model" {
		t.Fatalf("model file = %q, %v; want model", data, err)
	}

	backups, err := os.ReadDir(filepath.Join(parentDir, "Rime.backups"))
	if err != nil {
		t.Fatalf("read backup dir: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("backup count = %d; want 1", len(backups))
	}
	backupFile := filepath.Join(parentDir, "Rime.backups", backups[0].Name(), "default.custom.yaml")
	if data, err := os.ReadFile(backupFile); err != nil || string(data) != "old" {
		t.Fatalf("backup file = %q, %v; want old", data, err)
	}
}

func TestUpdateMainSchemeRejectsUnsafeZipAndRestoresBackup(t *testing.T) {
	parentDir := t.TempDir()
	targetDir := filepath.Join(parentDir, "Rime")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("create target dir: %v", err)
	}
	existingPath := filepath.Join(targetDir, "default.custom.yaml")
	if err := os.WriteFile(existingPath, []byte("old"), 0644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	err := UpdateMainScheme(testZip(t,
		zipEntry{name: "new.yaml", body: "new"},
		zipEntry{name: "../escape.yaml", body: "escape"},
	), targetDir)
	if err == nil {
		t.Fatal("UpdateMainScheme returned nil; want unsafe path error")
	}

	if data, err := os.ReadFile(existingPath); err != nil || string(data) != "old" {
		t.Fatalf("existing file after rollback = %q, %v; want old", data, err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "new.yaml")); !os.IsNotExist(err) {
		t.Fatalf("new file exists after rollback; stat error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(parentDir, "escape.yaml")); !os.IsNotExist(err) {
		t.Fatalf("escape file exists outside target; stat error: %v", err)
	}
}

type zipEntry struct {
	name string
	body string
}

func testZip(t *testing.T, entries ...zipEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, entry := range entries {
		w, err := zw.Create(entry.name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
