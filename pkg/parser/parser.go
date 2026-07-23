package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tokens  []Token
	current int
}

func Parse(input string) (*SelectStmt, error) {
	tokens, err := Lex(input)
	if err != nil {
		return nil, err
	}

	return NewParser(tokens).ParseSelect()
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) ParseSelect() (*SelectStmt, error) {
	if _, err := p.consume(TokenSelect, "se esperaba SELECT al inicio de la consulta"); err != nil {
		return nil, err
	}

	columnas, err := p.parseSelectColumns()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(TokenFrom, "se esperaba FROM despues de la lista de columnas"); err != nil {
		return nil, err
	}

	tabla, err := p.consume(TokenIdentifier, "se esperaba el nombre de la tabla despues de FROM")
	if err != nil {
		return nil, err
	}

	sentencia := &SelectStmt{
		Columns: columnas,
		Table:   tabla.Lexeme,
	}

	if p.match(TokenInner) {
		if _, err := p.consume(TokenJoin, "se esperaba JOIN después de INNER"); err != nil {
			return nil, err
		}
		joinTable, err := p.consume(TokenIdentifier, "se esperaba el nombre de la tabla después de JOIN")
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(TokenOn, "se esperaba ON después de JOIN"); err != nil {
			return nil, err
		}
		joinOn, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		sentencia.Join = &JoinClause{Table: joinTable.Lexeme, On: joinOn}
	}

	if p.match(TokenWhere) {
		where, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		sentencia.Where = where
	}

	// GROUP BY (opcional)
	if p.match(TokenGroup) {
		if _, err := p.consume(TokenBy, "se esperaba BY después de GROUP"); err != nil {
			return nil, err
		}
		grupos, err := p.parseListaColumnas()
		if err != nil {
			return nil, err
		}
		sentencia.GroupBy = grupos
	}

	// ORDER BY (opcional)
	if p.match(TokenOrder) {
		if _, err := p.consume(TokenBy, "se esperaba BY después de ORDER"); err != nil {
			return nil, err
		}
		orderBy, err := p.parseOrderByList()
		if err != nil {
			return nil, err
		}
		sentencia.OrderBy = orderBy
	}

	// LIMIT (opcional)
	if p.match(TokenLimit) {
		tokenNum, err := p.consume(TokenNumber, "se esperaba un número después de LIMIT")
		if err != nil {
			return nil, err
		}
		n, err := strconv.ParseInt(tokenNum.Lexeme, 10, 64)
		if err != nil {
			return nil, p.errorAt(tokenNum, "número inválido en LIMIT")
		}
		sentencia.Limit = &n
	}

	if _, err := p.consume(TokenEOF, "se esperaba el fin de la consulta"); err != nil {
		return nil, err
	}

	return sentencia, nil
}

func (p *Parser) parseSelectColumns() ([]SelectColumn, error) {
	if p.match(TokenAsterisk) {
		return []SelectColumn{{IsAsterisk: true}}, nil
	}

	primera, err := p.parseSelectColumn()
	if err != nil {
		return nil, err
	}

	columnas := []SelectColumn{primera}

	for p.match(TokenComma) {
		columna, err := p.parseSelectColumn()
		if err != nil {
			return nil, err
		}
		columnas = append(columnas, columna)
	}

	return columnas, nil
}

// parseSelectColumn parsea una columna del SELECT: un identificador o una función de agregación.
func (p *Parser) parseSelectColumn() (SelectColumn, error) {
	if p.esAggFunc() {
		agg, err := p.parseAggFunc()
		if err != nil {
			return SelectColumn{}, err
		}
		return SelectColumn{Agg: agg}, nil
	}

	name, err := p.parseQualifiedIdentifier()
	if err != nil {
		return SelectColumn{}, err
	}
	return SelectColumn{Name: name}, nil
}

// esAggFunc retorna true si el token actual es una función de agregación.
func (p *Parser) esAggFunc() bool {
	switch p.peek().Type {
	case TokenCount, TokenSum, TokenAvg, TokenMin, TokenMax:
		return true
	}
	return false
}

// parseAggFunc parsea COUNT(*), COUNT(col), SUM(col), AVG(col), MIN(col), MAX(col).
func (p *Parser) parseAggFunc() (*AggFunc, error) {
	funcToken := p.advance()
	nombre := strings.ToUpper(funcToken.Lexeme)

	if _, err := p.consume(TokenLeftParen, fmt.Sprintf("se esperaba ( después de %s", nombre)); err != nil {
		return nil, err
	}

	agg := &AggFunc{Name: nombre}

	if nombre == "COUNT" && p.match(TokenAsterisk) {
		agg.IsStar = true
	} else {
		col, err := p.consume(TokenIdentifier, fmt.Sprintf("se esperaba nombre de columna en %s()", nombre))
		if err != nil {
			return nil, err
		}
		agg.Column = col.Lexeme
	}

	if _, err := p.consume(TokenRightParen, fmt.Sprintf("se esperaba ) para cerrar %s()", nombre)); err != nil {
		return nil, err
	}

	return agg, nil
}

// parseListaColumnas parsea una lista de nombres de columna separados por coma.
func (p *Parser) parseListaColumnas() ([]string, error) {
	first, err := p.parseQualifiedIdentifier()
	if err != nil {
		return nil, err
	}
	cols := []string{first}
	for p.match(TokenComma) {
		col, err := p.parseQualifiedIdentifier()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)
	}
	return cols, nil
}

// parseOrderByList parsea la lista de criterios de ORDER BY.
func (p *Parser) parseOrderByList() ([]OrderByItem, error) {
	item, err := p.parseOrderByItem()
	if err != nil {
		return nil, err
	}
	items := []OrderByItem{item}
	for p.match(TokenComma) {
		item, err = p.parseOrderByItem()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// parseOrderByItem parsea un criterio de ordenamiento: nombre [ASC|DESC].
func (p *Parser) parseOrderByItem() (OrderByItem, error) {
	col, err := p.parseQualifiedIdentifier()
	if err != nil {
		return OrderByItem{}, err
	}
	desc := false
	if p.match(TokenDesc) {
		desc = true
	} else {
		p.match(TokenAsc) // ASC es opcional y se consume si está presente
	}
	return OrderByItem{Column: col, Desc: desc}, nil
}

func (p *Parser) parseExpression() (Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (Expr, error) {
	expr, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.match(TokenOr) {
		operador := p.previous()
		derecha, err := p.parseAnd()
		if err != nil {
			return nil, err
		}

		expr = &BinaryExpr{
			Left:     expr,
			Operator: strings.ToUpper(operador.Lexeme),
			Right:    derecha,
		}
	}

	return expr, nil
}

func (p *Parser) parseAnd() (Expr, error) {
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.match(TokenAnd) {
		operador := p.previous()
		derecha, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		expr = &BinaryExpr{
			Left:     expr,
			Operator: strings.ToUpper(operador.Lexeme),
			Right:    derecha,
		}
	}

	return expr, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if p.match(
		TokenEqual,
		TokenNotEqual,
		TokenLess,
		TokenGreater,
		TokenLessEqual,
		TokenGreatEqual,
	) {
		operador := p.previous()
		derecha, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		return &BinaryExpr{
			Left:     expr,
			Operator: operador.Lexeme,
			Right:    derecha,
		}, nil
	}

	return expr, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.match(TokenIdentifier) {
		name := p.previous().Lexeme
		if p.match(TokenDot) {
			col, err := p.consume(TokenIdentifier, "se esperaba nombre de columna después del punto")
			if err != nil {
				return nil, err
			}
			name = name + "." + col.Lexeme
		}
		return &Identifier{
			Name: name,
		}, nil
	}

	if p.match(TokenNumber) {
		token := p.previous()

		if strings.Contains(token.Lexeme, ".") {
			valor, err := strconv.ParseFloat(token.Lexeme, 64)
			if err != nil {
				return nil, p.errorAt(token, "numero decimal invalido")
			}

			return &Literal{
				Value: valor,
				Raw:   token.Lexeme,
			}, nil
		}

		valor, err := strconv.ParseInt(token.Lexeme, 10, 64)
		if err != nil {
			return nil, p.errorAt(token, "numero entero invalido")
		}

		return &Literal{
			Value: valor,
			Raw:   token.Lexeme,
		}, nil
	}

	if p.match(TokenString) {
		token := p.previous()
		contenido := strings.TrimSuffix(
			strings.TrimPrefix(token.Lexeme, "'"),
			"'",
		)

		return &Literal{
			Value: contenido,
			Raw:   token.Lexeme,
		}, nil
	}

	if p.match(TokenTrue) {
		return &Literal{
			Value: true,
			Raw:   p.previous().Lexeme,
		}, nil
	}

	if p.match(TokenFalse) {
		return &Literal{
			Value: false,
			Raw:   p.previous().Lexeme,
		}, nil
	}

	if p.match(TokenNull) {
		return &Literal{
			Value: nil,
			Raw:   p.previous().Lexeme,
		}, nil
	}

	if p.match(TokenLeftParen) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if _, err := p.consume(TokenRightParen, "se esperaba ) para cerrar la expresion"); err != nil {
			return nil, err
		}

		return expr, nil
	}

	return nil, p.errorAt(
		p.peek(),
		"se esperaba un identificador, literal o expresion entre parentesis",
	)
}

func (p *Parser) parseQualifiedIdentifier() (string, error) {
	first, err := p.consume(TokenIdentifier, "se esperaba un nombre de columna")
	if err != nil {
		return "", err
	}
	name := first.Lexeme
	if p.match(TokenDot) {
		second, err := p.consume(TokenIdentifier, "se esperaba nombre de columna después del punto")
		if err != nil {
			return "", err
		}
		name += "." + second.Lexeme
	}
	return name, nil
}

func (p *Parser) match(tipos ...TokenType) bool {
	for _, tipo := range tipos {
		if p.check(tipo) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) consume(tipo TokenType, mensaje string) (Token, error) {
	if p.check(tipo) {
		return p.advance(), nil
	}

	return Token{}, p.errorAt(p.peek(), mensaje)
}

func (p *Parser) check(tipo TokenType) bool {
	return p.peek().Type == tipo
}

func (p *Parser) advance() Token {
	token := p.peek()

	if p.current < len(p.tokens) {
		p.current++
	}

	return token
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == TokenEOF
}

func (p *Parser) peek() Token {
	if p.current >= len(p.tokens) {
		return Token{
			Type: TokenEOF,
			Position: Position{
				Line:   1,
				Column: 1,
			},
		}
	}

	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) errorAt(token Token, mensaje string) error {
	if token.Type == TokenEOF {
		mensaje = fmt.Sprintf("%s antes del fin de la consulta", mensaje)
	}

	return &SyntaxError{
		Message:  mensaje,
		Position: token.Position,
	}
}
