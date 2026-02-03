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

func TestParseWhereNameEquals(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.txt WHERE name = 'test.txt'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if command.WherePattern == nil {
		t.Fatal("WherePattern is nil")
	}

	if !command.WherePattern.MatchString("test.txt") {
		t.Error("pattern should match 'test.txt'")
	}

	if command.WherePattern.MatchString("other.txt") {
		t.Error("pattern should not match 'other.txt' (exact match)")
	}
}

func TestParseWhereNameLike(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT * FROM *.md WHERE name LIKE '%.tmp'")

	if command.Action != models.SELECT {
		t.Fatalf("expected SELECT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if command.WherePattern == nil {
		t.Fatal("WherePattern is nil")
	}

	if !command.WherePattern.MatchString("file.tmp") {
		t.Error("pattern should match 'file.tmp'")
	}

	if !command.WherePattern.MatchString("test.tmp") {
		t.Error("pattern should match 'test.tmp'")
	}

	if command.WherePattern.MatchString("file.md") {
		t.Error("pattern should not match 'file.md'")
	}
}

func TestParseCountWithWhereName(t *testing.T) {
	parser := mock.NewParser()
	command := parser.Parse("SELECT COUNT(*) FROM *.go WHERE name LIKE 'test%'")

	if command.Action != models.COUNT {
		t.Fatalf("expected COUNT, got %v", command.Action)
	}

	if command.WhereTarget != models.NAME {
		t.Fatalf("expected WhereTarget NAME, got %v", command.WhereTarget)
	}

	if !command.WherePattern.MatchString("test_file.go") {
		t.Error("pattern should match 'test_file.go'")
	}

	if command.WherePattern.MatchString("main.go") {
		t.Error("pattern should not match 'main.go'")
	}
}
