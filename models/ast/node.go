package ast

import (
	"github.com/overthinkinglabs/sqd/models"
)

type Node interface {
	Accept(visitor Visitor) (models.Command, error)
}
