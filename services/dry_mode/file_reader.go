package dry_mode

import (
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

type FileReader struct {
	errorHandler *ErrorHandler
	utils        *services.Utils
}

func NewFileReader(errorHandler *ErrorHandler, utils *services.Utils) *FileReader {
	return &FileReader{errorHandler: errorHandler, utils: utils}
}

func (fileReader *FileReader) ValidateAndReadFile(file string, stats *models.ExecutionStats) ([]string, bool) {
	if !fileReader.utils.IsPathInsideCwd(file) {
		fileReader.errorHandler.fail("invalid path: "+file, stats)
		return nil, false
	}

	if !fileReader.utils.CanWriteFile(file) {
		fileReader.errorHandler.fail("permission denied: "+file, stats)
		return nil, false
	}

	data, err := os.ReadFile(file)
	if err != nil {
		fileReader.errorHandler.fail(err.Error(), stats)
		return nil, false
	}

	return strings.Split(string(data), "\n"), true
}
