package mock

import (
	"github.com/albertoboccolini/sqd/services/sql"
)

func NewParser() *sql.Parser {
	extractor := sql.NewExtractor()
	whereParser := sql.NewWhereParser(extractor)
	batchParser := sql.NewBatchParser(extractor)
	parser := sql.NewParser(extractor, whereParser, batchParser)
	return parser
}
