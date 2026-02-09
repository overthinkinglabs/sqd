package ast

import (
	"github.com/albertoboccolini/sqd/models"
)

type Node interface {
	Accept(visitor Visitor) (models.Command, error)
}
