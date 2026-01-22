package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type SQLParser struct{}

func NewSQLParser() *SQLParser {
	return &SQLParser{}
}

func (sqlParser *SQLParser) Validate(sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return fmt.Errorf("Query cannot be empty")
	}

	upperSql := strings.ToUpper(sql)

	validPrefixes := []string{"SELECT COUNT", "SELECT", "UPDATE", "DELETE"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(upperSql, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("Query must start with SELECT, UPDATE, or DELETE")
	}

	if strings.HasPrefix(upperSql, "SELECT") && !strings.Contains(upperSql, "FROM") {
		return fmt.Errorf("SELECT query must contain FROM clause")
	}

	if strings.HasPrefix(upperSql, "UPDATE") && !strings.Contains(upperSql, "SET") {
		return fmt.Errorf("UPDATE query must contain SET clause")
	}

	if strings.HasPrefix(upperSql, "DELETE") && !strings.Contains(upperSql, "FROM") {
		return fmt.Errorf("DELETE query must contain FROM clause")
	}

	if strings.HasPrefix(upperSql, "UPDATE") {
		hasSetContent := strings.Contains(upperSql, "SET CONTENT")
		hasWhereName := strings.Contains(upperSql, "WHERE NAME")
		if hasSetContent && hasWhereName {
			return fmt.Errorf("SET content with WHERE name is not allowed")
		}
	}

	return nil
}

func (sqlParser *SQLParser) Parse(sql string) models.Command {
	sql = strings.TrimSpace(sql)
	upperSql := strings.ToUpper(sql)

	var command models.Command
	command.WhereTarget = models.Content
	command.SetTarget = models.Content
	command.SelectTarget = models.ALL

	if strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.COUNT
		command.File = sqlParser.extractBetween(sql, "FROM", "WHERE")
	}

	if strings.HasPrefix(upperSql, "SELECT") && !strings.HasPrefix(upperSql, "SELECT COUNT") {
		command.Action = models.SELECT
		command.File = sqlParser.extractBetween(sql, "FROM", "WHERE")
		command.SelectTarget = sqlParser.detectSelectTarget(upperSql)
	}

	if strings.HasPrefix(upperSql, "UPDATE") {
		command.Action = models.UPDATE
		command.File = sqlParser.extractBetween(sql, "UPDATE", "SET")
		command.SetTarget = sqlParser.detectSetTarget(upperSql)
	}

	if strings.HasPrefix(upperSql, "DELETE") {
		command.Action = models.DELETE
		command.File = sqlParser.extractBetween(sql, "DELETE FROM", "WHERE")
	}

	command.File = strings.TrimSpace(command.File)
	command.WhereTarget = sqlParser.detectWhereTarget(upperSql)

	if command.Action == models.UPDATE && strings.Count(upperSql, "SET CONTENT=") > 1 {
		command.IsBatch = true
		command.Replacements = sqlParser.parseBatchReplacements(sql)
		return command
	}

	if command.Action == models.DELETE && strings.Count(upperSql, "WHERE") > 1 {
		command.IsBatch = true
		command.Deletions = sqlParser.parseBatchDeletions(sql)
		return command
	}

	sqlParser.parseWhereClause(sql, &command)

	if command.Action == models.UPDATE {
		if command.SetTarget == models.Name {
			setNameRegex := regexp.MustCompile(`(?i)SET\s+name\s*=\s*['"]([^'"]*)['""]`)
			if matches := setNameRegex.FindStringSubmatch(sql); matches != nil {
				command.Replace = matches[1]
			}
		}
		if command.SetTarget == models.Content {
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
	}

	return command
}

func (sqlParser *SQLParser) detectSetTarget(upperSql string) models.Filter {
	if strings.Contains(upperSql, "SET NAME") {
		return models.Name
	}
	return models.Content
}

func (sqlParser *SQLParser) detectWhereTarget(upperSql string) models.Filter {
	if strings.Contains(upperSql, "WHERE NAME") {
		return models.Name
	}
	return models.Content
}

func (sqlParser *SQLParser) detectSelectTarget(upperSql string) models.Select {
	selectIdx := strings.Index(upperSql, "SELECT")
	fromIdx := strings.Index(upperSql, "FROM")
	if selectIdx == -1 || fromIdx == -1 {
		return models.ALL
	}

	selectClause := strings.TrimSpace(upperSql[selectIdx+6 : fromIdx])

	if selectClause == "NAME" {
		return models.NAME
	}
	if selectClause == "CONTENT" {
		return models.CONTENT
	}
	return models.ALL
}

func (sqlParser *SQLParser) parseWhereClause(sql string, command *models.Command) {
	filterField := "content"
	if command.WhereTarget == models.Name {
		filterField = "name"
	}

	exactRegex := regexp.MustCompile(`(?i)WHERE\s+` + filterField + `\s*=\s*`)
	likeRegex := regexp.MustCompile(`(?i)WHERE\s+` + filterField + `\s+LIKE\s*`)

	if loc := exactRegex.FindStringIndex(sql); loc != nil {
		command.MatchExact = true
		exactMatch := strings.TrimSpace(sql[loc[1]:])
		exactMatch = strings.Trim(exactMatch, " '\"")
		command.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		return
	}

	if loc := likeRegex.FindStringIndex(sql); loc != nil {
		command.MatchExact = false
		likePattern := sqlParser.extractAfter(sql[loc[1]:], "")
		likePattern = strings.Trim(likePattern, " '\"")
		command.Pattern = sqlParser.likeToRegex(likePattern)
	}
}

func (sqlParser *SQLParser) parseBatchDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)

		whereTarget := models.Content
		if strings.Contains(upper, "WHERE NAME") {
			whereTarget = models.Name
		}

		filterField := "content"
		if whereTarget == models.Name {
			filterField = "name"
		}

		exactPattern := `(?i)WHERE\s+` + filterField + `\s*=\s*`
		likePattern := `(?i)WHERE\s+` + filterField + `\s+LIKE\s*`

		if !regexp.MustCompile(exactPattern).MatchString(part) && !regexp.MustCompile(likePattern).MatchString(part) {
			continue
		}

		var del models.Deletion
		del.WhereTarget = whereTarget

		if regexp.MustCompile(exactPattern).MatchString(part) {
			del.MatchExact = true
			exactMatch := sqlParser.extractAfter(part, "WHERE "+filterField+" =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			del.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if regexp.MustCompile(likePattern).MatchString(part) {
			del.MatchExact = false
			likeValue := sqlParser.extractAfter(part, "LIKE")
			likeValue = strings.Trim(likeValue, " '\"")
			del.Pattern = sqlParser.likeToRegex(likeValue)
		}

		if del.Pattern != nil {
			deletions = append(deletions, del)
		}
	}

	return deletions
}

func (sqlParser *SQLParser) parseBatchReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if !strings.Contains(upperPart, "SET CONTENT=") && !strings.Contains(upperPart, "SET NAME=") {
			continue
		}

		var repl models.Replacement
		repl.SetTarget = models.Content
		repl.WhereTarget = models.Content

		if strings.Contains(upperPart, "SET NAME=") || strings.Contains(upperPart, "SET NAME =") {
			repl.SetTarget = models.Name
		}

		if strings.Contains(upperPart, "WHERE NAME") {
			repl.WhereTarget = models.Name
		}

		if repl.SetTarget == models.Content && repl.WhereTarget == models.Name {
			continue
		}

		if repl.SetTarget == models.Name {
			setNameRegex := regexp.MustCompile(`(?i)SET\s+name\s*=\s*'([^']*)'`)
			if matches := setNameRegex.FindStringSubmatch(part); matches != nil {
				repl.Replace = matches[1]
			}
		}
		if repl.SetTarget == models.Content {
			replaceValue := sqlParser.extractBetween(part, "SET content=", "WHERE")
			replaceValue = strings.Trim(replaceValue, " '\"")
			repl.Replace = replaceValue
		}

		filterField := "content"
		if repl.WhereTarget == models.Name {
			filterField = "name"
		}

		if strings.Contains(upperPart, "WHERE "+strings.ToUpper(filterField)+" =") {
			repl.MatchExact = true
			exactMatch := sqlParser.extractAfter(part, "WHERE "+filterField+" =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			repl.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperPart, "WHERE "+strings.ToUpper(filterField)+" LIKE") {
			repl.MatchExact = false
			likePattern := sqlParser.extractAfter(part, "LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			repl.Pattern = sqlParser.likeToRegex(likePattern)
		}

		replacements = append(replacements, repl)
	}

	return replacements
}

func (sqlParser *SQLParser) extractBetween(query, start, end string) string {
	upperStart := strings.ToUpper(start)
	upperEnd := strings.ToUpper(end)
	upperQuery := strings.ToUpper(query)

	startIndex := strings.Index(upperQuery, upperStart)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(upperStart)
	endIndex := strings.Index(upperQuery[startIndex:], upperEnd)

	if endIndex == -1 {
		return strings.TrimSpace(query[startIndex:])
	}

	return strings.TrimSpace(query[startIndex : startIndex+endIndex])
}

func (sqlParser *SQLParser) extractAfter(query, marker string) string {
	if marker == "" {
		return strings.TrimSpace(query)
	}

	markerUpper := strings.ToUpper(marker)
	upperQuery := strings.ToUpper(query)

	index := strings.Index(upperQuery, markerUpper)
	if index == -1 {
		return ""
	}

	return strings.TrimSpace(query[index+len(markerUpper):])
}

func (sqlParser *SQLParser) likeToRegex(pattern string) *regexp.Regexp {
	hasStart := strings.HasPrefix(pattern, "%")
	hasEnd := strings.HasSuffix(pattern, "%")

	if hasStart {
		pattern = pattern[1:]
	}

	if hasEnd {
		pattern = pattern[:len(pattern)-1]
	}

	hadTrailingSpace := hasEnd && strings.HasSuffix(pattern, " ")

	if hasEnd {
		pattern = strings.TrimRight(pattern, " ")
	}
	if hasStart {
		pattern = strings.TrimLeft(pattern, " ")
	}

	pattern = regexp.QuoteMeta(pattern)

	if !hasStart && hasEnd {
		if hadTrailingSpace {
			pattern = "^" + pattern + " "
		} else {
			pattern = "^" + pattern
		}
	}

	if hasStart && !hasEnd {
		pattern = pattern + "$"
	}

	if !hasStart && !hasEnd {
		pattern = "^" + pattern + "$"
	}

	return regexp.MustCompile(pattern)
}
