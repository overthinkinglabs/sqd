package displayable_errors

import "fmt"

type WalkError struct {
	*DisplayableError
}

func NewWalkError(directory string, err error) *WalkError {
	return &WalkError{
		DisplayableError: &DisplayableError{
			Description: fmt.Sprintf("Failed to walk directory %s: %v", directory, err),
		},
	}
}

func (walkError *WalkError) Unwrap() error {
	return walkError.DisplayableError
}
