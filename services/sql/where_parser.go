package sql

import (
	"regexp"
	"strings"

	"github.com/albertoboccolini/sqd/models"
)

type whereRule struct {
	Match  string
	Target models.WhereTarget
	After  string
	Negate bool
	Like   bool
}

var whereRules = []whereRule{
	{Match: "NAME !=", Target: models.WHERE_NAME, After: "NAME !=", Negate: true},
	{Match: "NAME LIKE", Target: models.WHERE_NAME, After: "NAME LIKE", Like: true},
	{Match: "NAME =", Target: models.WHERE_NAME, After: "NAME ="},
	{Match: "CONTENT !=", Target: models.WHERE_CONTENT, After: "CONTENT !=", Negate: true},
	{Match: "CONTENT LIKE", Target: models.WHERE_CONTENT, After: "CONTENT LIKE", Like: true},
	{Match: "CONTENT =", Target: models.WHERE_CONTENT, After: "CONTENT ="},
}

type WhereParser struct {
	extractor *Extractor
}

func NewWhereParser(extractor *Extractor) *WhereParser {
	return &WhereParser{extractor: extractor}
}

func (whereParser *WhereParser) splitCaseInsensitive(s, sep string) []string {
	upper := strings.ToUpper(s)
	sepUpper := strings.ToUpper(sep)
	var parts []string
	for {
		idx := strings.Index(upper, sepUpper)
		if idx == -1 {
			parts = append(parts, s)
			break
		}
		parts = append(parts, s[:idx])
		s = s[idx+len(sep):]
		upper = upper[idx+len(sepUpper):]
	}

	return parts
}

func (whereParser *WhereParser) parseCondition(part string) (models.WhereCondition, bool) {
	upperPart := strings.ToUpper(part)

	for _, rule := range whereRules {
		if !strings.Contains(upperPart, rule.Match) {
			continue
		}

		val := strings.Trim(whereParser.extractor.extractAfter(part, rule.After), " '\"")

		var pattern *regexp.Regexp
		if rule.Like {
			pattern = whereParser.extractor.likeToRegex(val)
		} else {
			pattern = regexp.MustCompile("^" + regexp.QuoteMeta(val) + "$")
		}

		return models.WhereCondition{
			Target:  rule.Target,
			Pattern: pattern,
			Negate:  rule.Negate,
		}, true
	}

	return models.WhereCondition{}, false
}

func (whereParser *WhereParser) Parse(sql string) ([]models.WhereCondition, models.WhereOperation) {
	upper := strings.ToUpper(sql)
	whereIdx := strings.Index(upper, "WHERE")
	if whereIdx == -1 {
		return nil, models.WHERE_SINGLE
	}

	wherePart := sql[whereIdx+5:]

	whereOperation := models.WHERE_OR
	var parts []string
	if strings.Contains(strings.ToUpper(wherePart), " AND ") {
		whereOperation = models.WHERE_AND
		parts = whereParser.splitCaseInsensitive(wherePart, " AND ")
	} else {
		parts = whereParser.splitCaseInsensitive(wherePart, " OR ")
	}

	var conditions []models.WhereCondition
	for _, part := range parts {
		condition, ok := whereParser.parseCondition(strings.TrimSpace(part))
		if ok {
			conditions = append(conditions, condition)
		}
	}

	return conditions, whereOperation
}
