package sql

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type Parser struct {
	extractor *Extractor
}

func NewParser(extractor *Extractor) *Parser {
	return &Parser{
		extractor: extractor,
	}
}

func (parser *Parser) parseBatchDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)

		if !strings.Contains(upper, "WHERE CONTENT =") {
			continue
		}

		var del models.Deletion
		del.MatchExact = true

		exactMatch := parser.extractor.extractAfter(part, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		del.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")

		deletions = append(deletions, del)
	}

	return deletions
}

func (parser *Parser) parseBatchReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if !strings.Contains(upperPart, "SET CONTENT=") {
			continue
		}

		var repl models.Replacement

		replaceValue := parser.extractor.extractBetween(part, "SET content=", "WHERE")
		replaceValue = strings.Trim(replaceValue, " '\"")
		repl.Replace = replaceValue

		if strings.Contains(upperPart, "WHERE CONTENT =") {
			repl.MatchExact = true
			exactMatch := parser.extractor.extractAfter(part, "WHERE content =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			repl.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperPart, "WHERE CONTENT LIKE") {
			repl.MatchExact = false
			likePattern := parser.extractor.extractAfter(part, "LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			repl.Pattern = parser.extractor.likeToRegex(likePattern)
		}

		replacements = append(replacements, repl)
	}

	return replacements
}

func (parser *Parser) detectSelectTarget(sql string) models.Select {
	upper := strings.ToUpper(sql)
	selectIdx := strings.Index(upper, "SELECT")
	fromIdx := strings.Index(upper, "FROM")
	if selectIdx == -1 || fromIdx == -1 {
		return models.ALL
	}

	selectClause := strings.TrimSpace(sql[selectIdx+6 : fromIdx])
	selectClauseLower := strings.ToLower(selectClause)

	if strings.HasPrefix(selectClauseLower, "count(") && strings.HasSuffix(selectClauseLower, ")") {
		selectClauseLower = strings.TrimSuffix(strings.TrimPrefix(selectClauseLower, "count("), ")")
	}

	switch selectClauseLower {
	case "name":
		return models.NAME
	case "content":
		return models.CONTENT
	default:
		return models.ALL
	}
}

func (parser *Parser) Parse(sql string) models.Command {
	sql = strings.TrimSpace(sql)
	upperSql := strings.ToUpper(sql)

	var command models.Command
	command.SelectTarget = models.ALL
	command.WhereTarget = models.WHERE_CONTENT

	if strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.COUNT
		command.File = parser.extractor.extractBetween(sql, "FROM", "WHERE")
		command.SelectTarget = parser.detectSelectTarget(sql)
	}

	if strings.HasPrefix(upperSql, "SELECT") && !strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.SELECT
		command.File = parser.extractor.extractBetween(sql, "FROM", "WHERE")
		command.SelectTarget = parser.detectSelectTarget(sql)
	}

	if strings.HasPrefix(upperSql, "UPDATE") {
		command.Action = models.UPDATE
		command.File = parser.extractor.extractBetween(sql, "UPDATE", "SET")
	}

	if strings.HasPrefix(upperSql, "DELETE") {
		command.Action = models.DELETE
		command.File = parser.extractor.extractBetween(sql, "DELETE FROM", "WHERE")
	}

	command.File = strings.TrimSpace(command.File)

	if command.Action == models.UPDATE && strings.Count(upperSql, "SET CONTENT=") > 1 {
		command.IsBatch = true
		command.Replacements = parser.parseBatchReplacements(sql)
		return command
	}

	if command.Action == models.DELETE && strings.Count(upperSql, "WHERE CONTENT =") > 1 {
		command.IsBatch = true
		command.Deletions = parser.parseBatchDeletions(sql)
		return command
	}

	if strings.Contains(upperSql, "WHERE NAME") {
		command.WhereTarget = models.WHERE_NAME

		if strings.Contains(upperSql, "WHERE NAME =") {
			exactMatch := parser.extractor.extractAfter(sql, "WHERE name =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			command.WherePattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperSql, "WHERE NAME LIKE") {
			likePattern := parser.extractor.extractAfter(sql, "WHERE name LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			command.WherePattern = parser.extractor.likeToRegex(likePattern)
		}

		return command
	}

	if strings.Contains(upperSql, "WHERE CONTENT") {
		whereContentRegex := regexp.MustCompile(`(?i)WHERE\s+content\s*=\s*`)
		loc := whereContentRegex.FindStringIndex(sql)
		if loc != nil {
			command.MatchExact = true
			exactMatch := strings.TrimSpace(sql[loc[1]:])
			exactMatch = strings.Trim(exactMatch, " '\"")
			command.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}
	}

	if strings.Contains(upperSql, "WHERE CONTENT LIKE") {
		command.MatchExact = false
		likePattern := parser.extractor.extractAfter(sql, "LIKE")
		likePattern = strings.Trim(likePattern, " '\"")
		command.Pattern = parser.extractor.likeToRegex(likePattern)
	}

	if command.Action == models.UPDATE {
		setContentRegex := regexp.MustCompile(`(?i)SET\s+content\s*=\s*`)
		whereIdx := strings.Index(upperSql, "WHERE")
		if whereIdx == -1 {
			whereIdx = len(sql)
		}

		loc := setContentRegex.FindStringIndex(sql)
		if loc != nil {
			command.Replace = strings.TrimSpace(sql[loc[1]:whereIdx])
			command.Replace = strings.Trim(command.Replace, "'\"")
		}
	}

	return command
}
