package parser

type Node interface {
	String() string
}

type Expr interface {
	Node
}

type SelectStmt struct {
	Columns []string
	Table   string
	Where   Expr
}
