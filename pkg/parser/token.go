package parser

import "fmt"

type TokenType string

const (
	TokenEOF        TokenType = "EOF"
	TokenIllegal    TokenType = "ILLEGAL"
	TokenIdentifier TokenType = "IDENTIFIER"
	TokenNumber     TokenType = "NUMBER"
	TokenString     TokenType = "STRING"
	TokenAsterisk   TokenType = "*"
	TokenComma      TokenType = ","
	TokenLeftParen  TokenType = "("
	TokenRightParen TokenType = ")"
	TokenEqual      TokenType = "="
	TokenNotEqual   TokenType = "<>"
	TokenLess       TokenType = "<"
	TokenGreater    TokenType = ">"
	TokenLessEqual  TokenType = "<="
	TokenGreatEqual TokenType = ">="
	TokenSelect     TokenType = "SELECT"
	TokenFrom       TokenType = "FROM"
	TokenWhere      TokenType = "WHERE"
	TokenAnd        TokenType = "AND"
	TokenOr         TokenType = "OR"
	TokenTrue       TokenType = "TRUE"
	TokenFalse      TokenType = "FALSE"
	TokenNull       TokenType = "NULL"
	// Cláusulas de ordenamiento y agrupación
	TokenOrder      TokenType = "ORDER"
	TokenBy         TokenType = "BY"
	TokenAsc        TokenType = "ASC"
	TokenDesc       TokenType = "DESC"
	TokenLimit      TokenType = "LIMIT"
	TokenGroup      TokenType = "GROUP"
	TokenHaving     TokenType = "HAVING"
	// Funciones de agregación
	TokenCount      TokenType = "COUNT"
	TokenSum        TokenType = "SUM"
	TokenAvg        TokenType = "AVG"
	TokenMin        TokenType = "MIN"
	TokenMax        TokenType = "MAX"
	// Otros
	TokenIs         TokenType = "IS"
	TokenNot        TokenType = "NOT"
)

type Position struct {
	Line   int
	Column int
	Offset int
}

func (p Position) String() string {
	return fmt.Sprintf("linea %d, columna %d", p.Line, p.Column)
}

type Token struct {
	Type     TokenType
	Lexeme   string
	Position Position
}

func (t Token) String() string {
	if t.Lexeme == "" {
		return string(t.Type)
	}

	return fmt.Sprintf("%s(%q)", t.Type, t.Lexeme)
}
