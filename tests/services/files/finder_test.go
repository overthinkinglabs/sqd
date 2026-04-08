package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/models/displayable_errors"
	"github.com/albertoboccolini/sqd/services/files"
)

func TestIsTextFileText(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("This is plain text\n")
	file.Close()

	finder := files.NewFinder()

	if !finder.IsTextFile(file.Name()) {
		t.Error("text file should be detected as text")
	}
}

func TestIsTextFileBinary(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.bin")
	defer os.Remove(file.Name())
	file.Write([]byte{0x00, 0x01, 0xFF, 0xFE, 0x00, 0x00})
	file.Close()

	finder := files.NewFinder()

	if finder.IsTextFile(file.Name()) {
		t.Error("binary file should not be detected as text")
	}
}

func TestIsTextFileNullByte(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.WriteString("text\x00more")
	file.Close()

	finder := files.NewFinder()

	if finder.IsTextFile(file.Name()) {
		t.Error("file with null byte should not be text")
	}
}

func TestIsTextFileControlChars(t *testing.T) {
	file, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(file.Name())
	file.Write([]byte{0x01, 0x02, 0x03})
	file.Close()

	finder := files.NewFinder()

	if finder.IsTextFile(file.Name()) {
		t.Error("file with control chars should not be text")
	}
}

func TestFindFilesWithPermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()

	fileA := filepath.Join(tmpDir, "file_a.md")
	os.WriteFile(fileA, []byte("content A"), 0o644)

	subdir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subdir, 0o755)

	fileBPath := filepath.Join(subdir, "file_b.md")
	os.WriteFile(fileBPath, []byte("content B"), 0o644)

	os.Chmod(subdir, 0o000)

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

	finder := files.NewFinder()
	foundFiles, err := finder.FindFiles("*.md")

	os.Chmod(subdir, 0o755)

	if len(foundFiles) == 0 {
		t.Errorf("should find file_a.md before permission denied on subdir")
	}

	if err == nil {
		t.Errorf("should return error with walk warnings")
		return
	}

	var errorCollection *models.ErrorCollection
	if !errors.As(err, &errorCollection) {
		t.Errorf("should return error collection with walk errors")
		return
	}

	if !errorCollection.HasErrors() {
		t.Errorf("error collection should have errors")
		return
	}

	var walkError *displayable_errors.WalkError
	found := false
	for _, e := range errorCollection.Errors() {
		if errors.As(e, &walkError) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("should contain walk error in error collection")
	}
}
