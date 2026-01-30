package sql

import (
	"regexp"
	"strings"
)

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (extractor *Extractor) extractBetween(query, start, end string) string {
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

func (extractor *Extractor) extractAfter(query, marker string) string {
	markerUpper := strings.ToUpper(marker)
	upperQuery := strings.ToUpper(query)

	index := strings.Index(upperQuery, markerUpper)
	if index == -1 {
		return ""
	}

	return strings.TrimSpace(query[index+len(markerUpper):])
}

func (extractor *Extractor) likeToRegex(pattern string) *regexp.Regexp {
	hasStart := strings.HasPrefix(pattern, "%")
	hasEnd := strings.HasSuffix(pattern, "%")

	if hasStart {
		pattern = pattern[1:]
	}

	if hasEnd && len(pattern) > 0 {
		pattern = pattern[:len(pattern)-1]
	}

	if pattern == "" {
		return regexp.MustCompile(".*")
	}

	pattern = regexp.QuoteMeta(pattern)

	if !hasStart && hasEnd {
		pattern = "^" + pattern
	}

	if hasStart && !hasEnd {
		pattern = pattern + "$"
	}

	if !hasStart && !hasEnd {
		pattern = "^" + pattern + "$"
	}

	return regexp.MustCompile(pattern)
}
