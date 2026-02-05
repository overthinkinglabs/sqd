package dry_mode

import (
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/models"
	"github.com/albertoboccolini/sqd/services"
)

type FileReader struct {
	utils *services.Utils
}

func NewFileReader(utils *services.Utils) *FileReader {
	return &FileReader{utils: utils}
}

func (fileReader *FileReader) fail(msg string, stats *models.ExecutionStats) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	stats.Skipped++
}

func (fileReader *FileReader) ValidateAndReadFile(file string, stats *models.ExecutionStats) ([]string, bool) {
	if !fileReader.utils.IsPathInsideCwd(file) {
		fileReader.fail("invalid path: "+file, stats)
		return nil, false
	}

	if !fileReader.utils.CanWriteFile(file) {
		fileReader.fail("permission denied: "+file, stats)
		return nil, false
	}

	data, err := os.ReadFile(file)
	if err != nil {
		fileReader.fail(err.Error(), stats)
		return nil, false
	}

	return strings.Split(string(data), "\n"), true
}
