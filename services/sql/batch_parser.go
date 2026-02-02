package sql

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type BatchParser struct {
	extractor *Extractor
}

func NewBatchParser(extractor *Extractor) *BatchParser {
	return &BatchParser{extractor: extractor}
}

func (batchParser *BatchParser) parseDeletions(sql string) []models.Deletion {
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

		exactMatch := batchParser.extractor.extractAfter(part, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")
		del.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")

		deletions = append(deletions, del)
	}

	return deletions
}

func (batchParser *BatchParser) parseReplacements(sql string) []models.Replacement {
	var replacements []models.Replacement

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upperPart := strings.ToUpper(part)

		if !strings.Contains(upperPart, "SET CONTENT=") {
			continue
		}

		var replacement models.Replacement

		replaceValue := batchParser.extractor.extractBetween(part, "SET content=", "WHERE")
		replaceValue = strings.Trim(replaceValue, " '\"")
		replacement.Replace = replaceValue

		if strings.Contains(upperPart, "WHERE CONTENT =") {
			replacement.MatchExact = true
			exactMatch := batchParser.extractor.extractAfter(part, "WHERE content =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			replacement.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperPart, "WHERE CONTENT LIKE") {
			replacement.MatchExact = false
			likePattern := batchParser.extractor.extractAfter(part, "LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			replacement.Pattern = batchParser.extractor.likeToRegex(likePattern)
		}

		replacements = append(replacements, replacement)
	}

	return replacements
}
