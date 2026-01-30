package mock

import (
	"github.com/albertoboccolini/sqd/services/sql"
)

func NewParser() *sql.Parser {
	extractor := sql.NewExtractor()
	parser := sql.NewParser(extractor)
	return parser
}
