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

	if p.match(TokenWhere) {
		where, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		sentencia.Where = where
	}

	if _, err := p.consume(TokenEOF, "se esperaba el fin de la consulta"); err != nil {
		return nil, err
	}

	return sentencia, nil
}

func (p *Parser) parseSelectColumns() ([]SelectColumn, error) {
	if p.match(TokenAsterisk) {
		return []SelectColumn{
			{
				IsAsterisk: true,
			},
		}, nil
	}

	primera, err := p.consume(TokenIdentifier, "se esperaba * o un nombre de columna despues de SELECT")
	if err != nil {
		return nil, err
	}

	columnas := []SelectColumn{
		{
			Name: primera.Lexeme,
		},
	}

	for p.match(TokenComma) {
		columna, err := p.consume(TokenIdentifier, "se esperaba un nombre de columna despues de la coma")
		if err != nil {
			return nil, err
		}

		columnas = append(columnas, SelectColumn{
			Name: columna.Lexeme,
		})
	}

	return columnas, nil
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
		return &Identifier{
			Name: p.previous().Lexeme,
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
