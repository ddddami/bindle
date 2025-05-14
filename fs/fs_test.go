package fs

import (
	"os"
	"testing"
)

func TestCreateDirIfNotExists(t *testing.T) {
	dir := "testdir"
	defer SafeRemove(dir)

	err := CreateDirIfNotExists(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !PathExists(dir) {
		t.Fatalf("expected path to exist, didn't")
	}
}

func TestCreateDirIfNotExistsWithPerm(t *testing.T) {
	dir := "testdir"
	defer SafeRemove(dir)

	err := CreateDirIfNotExistsWithPerm(dir, 0o700)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !PathExists(dir) {
		t.Fatalf("expected path to exist, didn't")
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if info.Mode().Perm() != 0o700 {
		t.Fatalf("expected permissions to be 0700, got %o", info.Mode().Perm())
	}
}

func TestPathExists(t *testing.T) {
	dir := "testdir"
	defer SafeRemove(dir)

	err := os.Mkdir(dir, 0o755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !PathExists(dir) {
		t.Fatalf("expected path to exist, didn't")
	}

	if PathExists("nonexistent") {
		t.Fatalf("expected path to not exist, did")
	}
}

func TestDirExists(t *testing.T) {
	dir := "testdir"
	defer SafeRemove(dir)

	err := os.Mkdir(dir, 0o755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !DirExists(dir) {
		t.Fatalf("expected path to exist, didn't")
	}

	if DirExists("nonexistent") {
		t.Fatalf("expected path to not exist, did")
	}
}

func TestListFilesWithExt(t *testing.T) {
	dir := "testdir"
	defer SafeRemove(dir)

	err := os.Mkdir(dir, 0o755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	files := []string{"file1.txt", "file2.txt", "file3.go"}
	for _, file := range files {
		var f *os.File
		f, err = os.Create(dir + "/" + file)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		f.Close()
	}

	gotFiles, err := ListFilesWithExt(dir, ".go")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(gotFiles) != 1 || gotFiles[0] != dir+"/file3.go" {
		t.Fatalf("expected one .go file, got %v", gotFiles)
	}
}

func TestSafeRemove(t *testing.T) {
	dir := "testdir"
	err := os.Mkdir(dir, 0o755)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = SafeRemove(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if PathExists(dir) {
		t.Fatalf("expected path to not exist, did")
	}
}

func TestCopyFile(t *testing.T) {
	src := "testsrc.txt"
	dst := "testdst.txt"
	defer SafeRemove(src)
	defer SafeRemove(dst)

	err := os.WriteFile(src, []byte("Hello, World!"), 0o644)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = CopyFile(src, dst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(data) != "Hello, World!" {
		t.Fatalf("expected file content to be 'Hello, World!', got %s", data)
	}
}
