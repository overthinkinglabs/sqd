package tests

import (
	"testing"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/tests/mock"
)

func TestParseSelect(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM test.txt WHERE content LIKE '%foo%'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.File != "test.txt" {
		t.Fatalf("expected test.txt, got %s", command.File)
	}

	if command.Pattern == nil {
		t.Fatal("pattern is nil")
	}

	if !command.Pattern.MatchString("foo") {
		t.Error("pattern should match 'foo'")
	}
}

func TestParseCount(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM file.sql WHERE content = 'exact'")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if command.File != "file.sql" {
		t.Fatalf("expected file.sql, got %s", command.File)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}
}

func TestParseUpdate(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("UPDATE file.txt SET content='new' WHERE content = 'old'")

	if command.Action != models.UPDATE {
		t.Fatalf("expected UPDATE, got %v", command.Action)
	}

	if command.Replace != "new" {
		t.Fatalf("expected 'new', got %s", command.Replace)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}

	if !command.Pattern.MatchString("old") {
		t.Error("pattern should match 'old'")
	}
}

func TestParseDelete(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("DELETE FROM file.txt WHERE content = 'remove'")

	if command.Action != models.DELETE {
		t.Fatalf("expected DELETE, got %v", command.Action)
	}

	if !command.MatchExact {
		t.Error("expected exact match")
	}
}

func TestExactMatch(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM f WHERE content = 'exact'")

	if !command.Pattern.MatchString("exact") {
		t.Error("should match 'exact'")
	}

	if command.Pattern.MatchString("not exact") {
		t.Error("should not match 'not exact'")
	}
}

func TestUpdatePreservesSpace(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("UPDATE *.md SET content = '#### ' WHERE content LIKE '### %'")

	input := "### Title"
	result := command.Pattern.ReplaceAllString(input, command.Replace)

	if result != "#### Title" {
		t.Errorf("expected '#### Title', got '%s'", result)
	}
}
