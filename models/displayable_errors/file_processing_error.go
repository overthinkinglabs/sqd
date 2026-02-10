package displayable_errors

import "fmt"

type FileProcessingError struct {
	DisplayableError
}

func NewFileProcessingError(filename string, err error) *DisplayableError {
	return &DisplayableError{
		Description: fmt.Sprintf("Failed to process file %s: %v", filename, err),
	}
}
