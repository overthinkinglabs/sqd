package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/tests/mock"
)

func TestValidateTransactionModeStopsOnFirstError(t *testing.T) {
	dryRunner := mock.NewDryModeRunner()

	pattern := regexp.MustCompile("test")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}
	cwd, _ := os.Getwd()
	validFile := filepath.Join(cwd, "valid.txt")
	invalidFile := filepath.Join(cwd, "..", "invalid.txt")

	result := dryRunner.Validate(command, []string{invalidFile, validFile}, stats, true, false)

	if result {
		t.Error("expected transaction mode to return false on error")
	}
}

func TestValidateNonTransactionModeContinuesAfterError(t *testing.T) {
	cwd, _ := os.Getwd()
	validFile := filepath.Join(cwd, "valid.txt")
	os.WriteFile(validFile, []byte("content\n"), 0644)
	defer os.Remove(validFile)

	dryRunner := mock.NewDryModeRunner()
	pattern := regexp.MustCompile("content")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}
	invalidFile := filepath.Join(cwd, "..", "invalid.txt")

	result := dryRunner.Validate(command, []string{invalidFile, validFile}, stats, false, false)

	if !result {
		t.Error("expected non-transaction mode to return true")
	}
}

func TestValidatePermissionDenied(t *testing.T) {
	cwd, _ := os.Getwd()
	testFile := filepath.Join(cwd, "readonly.txt")
	os.WriteFile(testFile, []byte("content\n"), 0400)
	defer os.Remove(testFile)

	dryRunner := mock.NewDryModeRunner()

	pattern := regexp.MustCompile("content")
	command := models.Command{
		Action:  models.UPDATE,
		Pattern: pattern,
		Replace: "changed",
	}

	stats := &models.ExecutionStats{}

	dryRunner.Validate(command, []string{testFile}, stats, false, false)

	if stats.Skipped != 1 {
		t.Errorf("expected 1 skipped for permission denied, got %d", stats.Skipped)
	}

	if stats.Processed != 0 {
		t.Errorf("expected 0 processed for permission denied, got %d", stats.Processed)
	}
}
