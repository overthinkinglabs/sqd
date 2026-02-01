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

func (batchParser *BatchParser) ParseDeletions(sql string) []models.Deletion {
	var deletions []models.Deletion

	parts := strings.SplitSeq(sql, ",")

	for part := range parts {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)

		if !strings.Contains(upper, "WHERE CONTENT =") {
			continue
		}

		exactMatch := batchParser.extractor.extractAfter(part, "WHERE content =")
		exactMatch = strings.Trim(exactMatch, " '\"")

		deletions = append(deletions, models.Deletion{
			MatchExact: true,
			Pattern:    regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$"),
		})
	}

	return deletions
}

func (batchParser *BatchParser) ParseReplacements(sql string) []models.Replacement {
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
		replacement.Replace = strings.Trim(replaceValue, " '\"")

		if strings.Contains(upperPart, "WHERE CONTENT =") {
			replacement.MatchExact = true
			exactMatch := batchParser.extractor.extractAfter(part, "WHERE content =")
			exactMatch = strings.Trim(exactMatch, " '\"")
			replacement.Pattern = regexp.MustCompile("^" + regexp.QuoteMeta(exactMatch) + "$")
		}

		if strings.Contains(upperPart, "WHERE CONTENT LIKE") {
			likePattern := batchParser.extractor.extractAfter(part, "LIKE")
			likePattern = strings.Trim(likePattern, " '\"")
			replacement.Pattern = batchParser.extractor.likeToRegex(likePattern)
		}

		replacements = append(replacements, replacement)
	}

	return replacements
}
