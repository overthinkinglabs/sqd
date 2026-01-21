package files

import (
	"fmt"
	"os"
	"strings"

	"github.com/albertoboccolini/sqd/services"
)

type Processor struct {
	utils *services.Utils
}

func NewProcessor(utils *services.Utils) *Processor {
	return &Processor{utils: utils}
}

func (processor *Processor) ProcessFile(filename string, transformFunc func([]string) ([]string, int)) (int, error) {
	if !processor.utils.IsPathInsideCwd(filename) {
		return 0, fmt.Errorf("invalid path detected: %s", filename)
	}
	if !processor.utils.CanWriteFile(filename) {
		return 0, fmt.Errorf("permission denied")
	}

	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	newLines, count := transformFunc(lines)

	if count > 0 {
		err = os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), info.Mode().Perm())
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}
