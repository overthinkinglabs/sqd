package displayable_errors

import "fmt"

type FileTooLargeError struct {
	*DisplayableError
}

func NewFileTooLargeError(filename string, size int64, maxSize int64) *FileTooLargeError {
	return &FileTooLargeError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("File %s is too large (%d bytes, max %d bytes)", filename, size, maxSize),
		},
	}
}

func (fileTooLargeError *FileTooLargeError) Unwrap() error {
	return fileTooLargeError.DisplayableError
}
