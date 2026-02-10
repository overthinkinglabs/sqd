package models

type ErrorCollection struct {
	errors []error
}

func NewErrorCollection() *ErrorCollection {
	return &ErrorCollection{
		errors: make([]error, 0),
	}
}

func (errorCollection *ErrorCollection) Add(err error) {
	if err != nil {
		errorCollection.errors = append(errorCollection.errors, err)
	}
}

func (errorCollection *ErrorCollection) HasErrors() bool {
	return len(errorCollection.errors) > 0
}

func (errorCollection *ErrorCollection) Errors() []error {
	return errorCollection.errors
}

func (errorCollection *ErrorCollection) Error() string {
	if !errorCollection.HasErrors() {
		return ""
	}

	result := ""
	for _, err := range errorCollection.errors {
		if result != "" {
			result += "\n"
		}
		result += err.Error()
	}
	return result
}
