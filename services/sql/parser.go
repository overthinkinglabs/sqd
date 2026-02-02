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

func (parser *Parser) detectSelectTarget(sql string) models.TokenType {
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
		command.SelectTarget = parser.detectSelectTarget(sql)
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
		whereCount := strings.Count(upperSql, "WHERE")
		if whereCount > 1 {
			command.IsBatch = true
			command.File = parser.extractor.extractFilename(sql, "DELETE FROM", "WHERE")
			command.Deletions = parser.batchParser.parseDeletions(sql)
			return command
		}
	}

	for !parser.currentTokenIs(models.EOF) {
		if parser.currentTokenIs(models.FROM) {
			parser.nextToken()
			command.File = parser.currentToken.Literal
		}

		if parser.currentTokenIs(models.WHERE) {
			parser.nextToken()

			if parser.currentTokenIs(models.NAME) {
				command.WhereTarget = models.NAME
				parser.nextToken()

				if parser.currentTokenIs(models.EQUALS) {
					parser.nextToken()
					exactMatch := parser.currentToken.Literal
					command.WherePattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
				}

				if parser.currentTokenIs(models.LIKE) {
					parser.nextToken()
					likePattern := parser.currentToken.Literal
					command.WherePattern = parser.extractor.likeToRegex(likePattern)
				}
			}

			if parser.currentTokenIs(models.CONTENT) {
				parser.nextToken()

				if parser.currentTokenIs(models.EQUALS) {
					parser.nextToken()
					exactMatch := parser.currentToken.Literal
					command.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
				}

				if parser.currentTokenIs(models.LIKE) {
					parser.nextToken()
					likePattern := parser.currentToken.Literal
					command.Pattern = parser.extractor.likeToRegex(likePattern)
				}
			}
		}

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

		parser.nextToken()
	}

	command.File = strings.TrimSpace(command.File)
	return command
}
