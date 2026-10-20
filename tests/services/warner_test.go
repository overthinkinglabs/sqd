package tests

import (
	"strings"
	"testing"

	"github.com/overthinkinglabs/sqd/src/services"
)

func TestWarnerConfirmAcceptsYes(t *testing.T) {
	answers := []string{"y\n", "Y\n", "yes\n", "YES\n"}

	for _, answer := range answers {
		warner := services.NewWarner(strings.NewReader(answer))

		proceed, err := warner.Confirm("test message")
		if err != nil {
			t.Fatalf("unexpected error for answer %q: %v", answer, err)
		}

		if !proceed {
			t.Errorf("expected proceed for answer %q", answer)
		}
	}
}

func TestWarnerConfirmAbortsOnNo(t *testing.T) {
	answers := []string{"n\n", "no\n", "maybe\n", "\n", ""}

	for _, answer := range answers {
		warner := services.NewWarner(strings.NewReader(answer))

		proceed, err := warner.Confirm("test message")
		if err != nil {
			t.Fatalf("unexpected error for answer %q: %v", answer, err)
		}

		if proceed {
			t.Errorf("expected abort for answer %q", answer)
		}
	}
}
