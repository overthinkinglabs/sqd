package sql

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type Parser struct {
	extractor    *Extractor
	batchParser  *BatchParser
	lexer        *Lexer
	currentToken models.Token
	peekToken    models.Token
}

func NewParser(extractor *Extractor, batchParser *BatchParser) *Parser {
	return &Parser{
		extractor:   extractor,
		batchParser: batchParser,
	}
}

func (parser *Parser) initLexer(input string) {
	parser.lexer = NewLexer(input)
	parser.nextToken()
	parser.nextToken()
}

func (parser *Parser) nextToken() {
	parser.currentToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
}

func (parser *Parser) currentTokenIs(tokenType models.TokenType) bool {
	return parser.currentToken.Type == tokenType
}

func (parser *Parser) peekTokenIs(tokenType models.TokenType) bool {
	return parser.peekToken.Type == tokenType
}

func (parser *Parser) parseSelectTarget(sql string) models.TokenType {
	lexer := NewLexer(sql)

	for {
		token := lexer.NextToken()

		if token.Type == models.SELECT {
			token = lexer.NextToken()

			if token.Type == models.COUNT {
				token = lexer.NextToken()
				if token.Type == models.LPAREN {
					token = lexer.NextToken()
				}
			}

			if token.Type == models.NAME {
				return models.NAME
			}

			if token.Type == models.CONTENT {
				return models.CONTENT
			}

			if token.Type == models.ASTERISK {
				return models.ASTERISK
			}
		}

		if token.Type == models.EOF {
			break
		}
	}

	return models.ASTERISK
}

func (parser *Parser) parseComparison(pattern **regexp.Regexp, negate *bool) {
	parser.nextToken()

	if parser.currentTokenIs(models.EQUALS) {
		parser.nextToken()
		exactMatch := parser.currentToken.Literal
		*pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		*negate = false
		return
	}

	if parser.currentTokenIs(models.NOT_EQUALS) {
		parser.nextToken()
		exactMatch := parser.currentToken.Literal
		*pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		*negate = true
		return
	}

	if parser.currentTokenIs(models.LIKE) {
		parser.nextToken()
		likePattern := parser.currentToken.Literal
		*pattern = parser.extractor.likeToRegex(likePattern)
		*negate = false
	}
}

func (parser *Parser) parseWhereClause(command *models.Command) {
	if !parser.currentTokenIs(models.WHERE) {
		return
	}

	parser.nextToken()

	if parser.currentTokenIs(models.NAME) {
		command.WhereTarget = models.NAME
		parser.parseComparison(&command.WherePattern, &command.NegateFileName)
		return
	}

	if parser.currentTokenIs(models.CONTENT) {
		parser.parseComparison(&command.Pattern, &command.NegateContent)
	}
}

func (parser *Parser) parseOrderItem() models.OrderBy {
	item := models.OrderBy{}

	switch {
	case parser.currentTokenIs(models.NAME):
		item.Column = models.NAME
	case parser.currentTokenIs(models.CONTENT):
		item.Column = models.CONTENT
	default:
		return item
	}

	parser.nextToken()

	switch {
	case parser.currentTokenIs(models.ASC):
		item.Direction = models.ASC
		parser.nextToken()
	case parser.currentTokenIs(models.DESC):
		item.Direction = models.DESC
		parser.nextToken()
	}

	return item
}

func (parser *Parser) parseOrderBy(command *models.Command) {
	if !parser.currentTokenIs(models.ORDER) || !parser.peekTokenIs(models.BY) {
		return
	}

	parser.nextToken()
	parser.nextToken()

	command.OrderBy = make([]models.OrderBy, 0)

	for {
		orderItem := parser.parseOrderItem()
		if orderItem.Column == models.TokenType(0) {
			break
		}

		command.OrderBy = append(command.OrderBy, orderItem)

		if !parser.currentTokenIs(models.COMMA) {
			break
		}
		parser.nextToken()
	}
}

func (parser *Parser) Parse(sql string) models.Command {
	parser.initLexer(sql)

	var command models.Command
	command.SelectTarget = models.ASTERISK
	command.WhereTarget = models.CONTENT

	upperSql := strings.ToUpper(sql)

	if parser.currentTokenIs(models.SELECT) {
		if parser.peekTokenIs(models.COUNT) {
			command.Action = models.COUNT
			parser.nextToken()
		}

		if command.Action != models.COUNT {
			command.Action = models.SELECT
		}
		command.SelectTarget = parser.parseSelectTarget(sql)
	}

	if parser.currentTokenIs(models.UPDATE) {
		command.Action = models.UPDATE
		parser.nextToken()
		if parser.currentTokenIs(models.IDENTIFIER) {
			command.File = parser.currentToken.Literal
		}
	}

	if parser.currentTokenIs(models.DELETE) {
		command.Action = models.DELETE
	}

	if command.Action == models.UPDATE {
		setCount := strings.Count(upperSql, "SET")
		if setCount > 1 {
			command.IsBatch = true
			command.File = parser.extractor.extractFilename(sql, "UPDATE", "SET")
			command.Replacements = parser.batchParser.parseBatchReplacements(sql)
			return command
		}
	}

	if command.Action == models.DELETE {
		whereIdx := strings.Index(upperSql, "WHERE")
		if whereIdx != -1 {
			whereClause := sql[whereIdx+5:]
			commaCount := strings.Count(whereClause, ",")

			if commaCount > 0 {
				command.IsBatch = true
				command.File = parser.extractor.extractFilename(sql, "DELETE FROM", "WHERE")
				command.Deletions = parser.batchParser.parseDeletions(sql)
				return command
			}
		}
	}

	for !parser.currentTokenIs(models.EOF) {
		if parser.currentTokenIs(models.FROM) {
			parser.nextToken()
			command.File = parser.currentToken.Literal
		}

		parser.parseWhereClause(&command)

		if parser.currentTokenIs(models.SET) && command.Action == models.UPDATE {
			parser.nextToken()

			if parser.currentTokenIs(models.CONTENT) {
				parser.nextToken()
				if parser.currentTokenIs(models.EQUALS) {
					parser.nextToken()
					command.Replace = parser.currentToken.Literal
				}
			}
		}

		parser.parseOrderBy(&command)
		parser.nextToken()
	}

	command.File = strings.TrimSpace(command.File)
	return command
}
