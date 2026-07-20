package parser

import (
	"fmt"
	"strings"
)

type Node interface {
	String() string
}

type Expr interface {
	Node
	exprNode()
}

type SelectStmt struct {
	Columns []SelectColumn
	Table   string
	Where   Expr
}

func (s *SelectStmt) String() string {
	if s == nil {
		return ""
	}

	columnas := make([]string, len(s.Columns))
	for indice, columna := range s.Columns {
		columnas[indice] = columna.String()
	}

	consulta := fmt.Sprintf(
		"SELECT %s FROM %s",
		strings.Join(columnas, ", "),
		s.Table,
	)

	if s.Where != nil {
		consulta += " WHERE " + s.Where.String()
	}

	return consulta
}

type SelectColumn struct {
	Name       string
	IsAsterisk bool
}

func (c SelectColumn) String() string {
	if c.IsAsterisk {
		return "*"
	}

	return c.Name
}

type Identifier struct {
	Name string
}

func (i *Identifier) exprNode() {}

func (i *Identifier) String() string {
	if i == nil {
		return ""
	}

	return i.Name
}

type Literal struct {
	Value any
	Raw   string
}

func (l *Literal) exprNode() {}

func (l *Literal) String() string {
	if l == nil {
		return ""
	}

	if l.Raw != "" {
		return l.Raw
	}

	return fmt.Sprint(l.Value)
}

type BinaryExpr struct {
	Left     Expr
	Operator string
	Right    Expr
}

func (b *BinaryExpr) exprNode() {}

func (b *BinaryExpr) String() string {
	if b == nil {
		return ""
	}

	return fmt.Sprintf(
		"(%s %s %s)",
		b.Left.String(),
		b.Operator,
		b.Right.String(),
	)
}
