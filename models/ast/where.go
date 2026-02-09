package ast

import (
	"regexp"

	"github.com/albertoboccolini/sqd/models"
)

type Where struct {
	Target  models.TokenType
	Pattern *regexp.Regexp
	Negate  bool
}
