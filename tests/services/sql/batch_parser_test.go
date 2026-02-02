package tests

import (
	"testing"

	"github.com/albertoboccolini/sqd/tests/mock"
)

func TestParseBatchUpdate(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("UPDATE file.txt SET content='a' WHERE content = 'x', SET content='b' WHERE content = 'y'")

	if !command.IsBatch {
		t.Fatal("expected batch mode")
	}

	if len(command.Replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %d", len(command.Replacements))
	}

	if command.Replacements[0].Replace != "a" {
		t.Errorf("first replacement: expected 'a', got %s", command.Replacements[0].Replace)
	}

	if command.Replacements[1].Replace != "b" {
		t.Errorf("second replacement: expected 'b', got %s", command.Replacements[1].Replace)
	}
}

func TestParseBatchDelete(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("DELETE FROM file.txt WHERE content = 'x', WHERE content = 'y'")

	if !command.IsBatch {
		t.Fatal("expected batch mode")
	}

	if len(command.Deletions) != 2 {
		t.Fatalf("expected 2 deletions, got %d", len(command.Deletions))
	}
}
