package services

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	yellowEscape = "\033[1;33m"
	resetEscape  = "\033[0m"
)

type Warner struct {
	input io.Reader
}

func NewWarner(input io.Reader) *Warner {
	return &Warner{input: input}
}

func (warner *Warner) Warn(message string) {
	fmt.Fprintf(os.Stderr, "%sWARNING: %s%s\n", yellowEscape, message, resetEscape)
}

func (warner *Warner) Confirm(message string) (bool, error) {
	warner.Warn(message)
	fmt.Fprintf(os.Stderr, "%sProceed? [y/N]: %s", yellowEscape, resetEscape)

	answer, readErr := bufio.NewReader(warner.input).ReadString('\n')
	if readErr != nil && readErr != io.EOF {
		return false, readErr
	}

	normalized := strings.ToLower(strings.TrimSpace(answer))
	return normalized == "y" || normalized == "yes", nil
}
