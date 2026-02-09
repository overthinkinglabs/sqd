package ast

import "github.com/albertoboccolini/sqd/models"

type UpdateStatement struct {
	Source       string
	Replacements []models.Replacement
	WhereClause  *Where
	IsBatch      bool
}

func (statement *UpdateStatement) Accept(visitor Visitor) (models.Command, error) {
	return visitor.VisitUpdate(statement)
}
