package mock

import (
	"github.com/albertoboccolini/sqd/services/sql"
)

func NewParser() *sql.Parser {
	extractor := sql.NewExtractor()
	batchParser := sql.NewBatchParser(extractor)
	commandBuilder := sql.NewCommandBuilder()
	return sql.NewParser(extractor, batchParser, commandBuilder)
}
