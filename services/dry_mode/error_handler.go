package dry_mode

import (
	"fmt"
	"os"

	"github.com/albertoboccolini/sqd/models"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (errorHandler *ErrorHandler) fail(msg string, stats *models.ExecutionStats) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	stats.Skipped++
}
