package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/tests/mock"
)

func TestParallelProcessingWithErrors(t *testing.T) {
	cwd, _ := os.Getwd()

	file1 := filepath.Join(cwd, "parallel_valid1.txt")
	file2 := filepath.Join(cwd, "parallel_valid2.txt")
	invalidFile := filepath.Join(cwd, "nonexistent_file.txt")

	os.WriteFile(file1, []byte("content1\ncontent1\n"), 0644)
	os.WriteFile(file2, []byte("content2\ncontent2\ncontent2\n"), 0644)
	defer os.Remove(file1)
	defer os.Remove(file2)

	files := []string{file1, invalidFile, file2}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("COUNT parallel_*.txt WHERE content LIKE 'content'")

	dispatcher.Execute(command, files, false, false, false)

	if _, err := os.Stat(file1); err != nil {
		t.Error("valid file1 should still exist after processing")
	}

	if _, err := os.Stat(file2); err != nil {
		t.Error("valid file2 should still exist after processing")
	}

	if _, err := os.Stat(invalidFile); err == nil {
		t.Error("invalid file should not exist")
	}
}

func TestParallelProcessingConcurrencyLimit(t *testing.T) {
	cwd, _ := os.Getwd()

	numFiles := models.MAX_CONCURRENT_GOROUTINES * 2
	files := make([]string, numFiles)

	for i := range numFiles {
		file := filepath.Join(cwd, "concurrent_test_"+string(rune(i%26+'a'))+string(rune(i/26+'0'))+".txt")
		os.WriteFile(file, []byte("test line\n"), 0644)
		files[i] = file
		defer os.Remove(file)
	}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("COUNT concurrent_test_*.txt WHERE content LIKE 'test'")

	dispatcher.Execute(command, files, false, false, false)

	for i, file := range files {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("file %d should exist after high concurrency processing: %v", i, err)
		}
	}
}

func TestParallelProcessingMaintainsFileIntegrity(t *testing.T) {
	cwd, _ := os.Getwd()

	numFiles := models.MAX_CONCURRENT_GOROUTINES * 2
	files := make([]string, numFiles)
	expectedContent := make(map[string]string)

	for i := range numFiles {
		file := filepath.Join(cwd, "integrity_test_"+string(rune('a'+i))+".txt")
		content := strings.Repeat("line "+string(rune('0'+i))+"\n", i+1)
		os.WriteFile(file, []byte(content), 0644)
		files[i] = file
		expectedContent[file] = content
		defer os.Remove(file)
	}

	dispatcher := mock.NewDispatcher()
	parser := mock.NewParser()
	command := parser.Parse("COUNT integrity_test_*.txt WHERE content LIKE 'line'")

	dispatcher.Execute(command, files, false, false, false)

	for file, expected := range expectedContent {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Errorf("failed to read file %s: %v", file, err)
			continue
		}

		if string(data) != expected {
			t.Errorf("file %s content corrupted during parallel processing", file)
		}
	}
}
