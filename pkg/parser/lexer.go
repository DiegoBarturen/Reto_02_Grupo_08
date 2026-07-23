package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type Lexer struct {
	input  []rune
	start  int
	actual int
	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  []rune(input),
		line:   1,
		column: 1,
	}
}

func Lex(input string) ([]Token, error) {
	return NewLexer(input).Tokens()
}

func (l *Lexer) Tokens() ([]Token, error) {
	tokens := make([]Token, 0)

	for {
		l.skipWhitespace()
		l.start = l.actual
		posicion := l.position()

		if l.isAtEnd() {
			tokens = append(tokens, Token{
				Type:     TokenEOF,
				Position: posicion,
			})
			return tokens, nil
		}

		caracter := l.advance()

		switch caracter {
		case '*':
			tokens = append(tokens, l.token(TokenAsterisk, posicion))
		case ',':
			tokens = append(tokens, l.token(TokenComma, posicion))
<<<<<<< HEAD
		case '.':
			tokens = append(tokens, l.token(TokenDot, posicion))
=======
>>>>>>> af2c9a5137fac5ac5ffaed2e81ebc59fd20fca5a
		case '(':
			tokens = append(tokens, l.token(TokenLeftParen, posicion))
		case ')':
			tokens = append(tokens, l.token(TokenRightParen, posicion))
		case '=':
			tokens = append(tokens, l.token(TokenEqual, posicion))
		case '<':
			if l.match('=') {
				tokens = append(tokens, l.token(TokenLessEqual, posicion))
			} else if l.match('>') {
				tokens = append(tokens, l.token(TokenNotEqual, posicion))
			} else {
				tokens = append(tokens, l.token(TokenLess, posicion))
			}
		case '>':
			if l.match('=') {
				tokens = append(tokens, l.token(TokenGreatEqual, posicion))
			} else {
				tokens = append(tokens, l.token(TokenGreater, posicion))
			}
		case '\'':
			token, err := l.stringToken(posicion)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		default:
			if isIdentifierStart(caracter) {
				tokens = append(tokens, l.identifier(posicion))
				continue
			}

			if unicode.IsDigit(caracter) {
				tokens = append(tokens, l.number(posicion))
				continue
			}

			return nil, &SyntaxError{
				Message:  fmt.Sprintf("caracter no reconocido %q", caracter),
				Position: posicion,
			}
		}
	}
}

func (l *Lexer) skipWhitespace() {
	for !l.isAtEnd() {
		switch l.peek() {
		case ' ', '\t', '\r':
			l.advance()
		case '\n':
			l.advance()
		default:
			return
		}
	}
}

func (l *Lexer) identifier(posicion Position) Token {
	for !l.isAtEnd() && isIdentifierPart(l.peek()) {
		l.advance()
	}

	lexema := l.lexeme()
	tipo := keywordTokenType(lexema)

	return Token{
		Type:     tipo,
		Lexeme:   lexema,
		Position: posicion,
	}
}

func (l *Lexer) number(posicion Position) Token {
	for !l.isAtEnd() && unicode.IsDigit(l.peek()) {
		l.advance()
	}

	if !l.isAtEnd() && l.peek() == '.' && l.peekNextIsDigit() {
		l.advance()

		for !l.isAtEnd() && unicode.IsDigit(l.peek()) {
			l.advance()
		}
	}

	return Token{
		Type:     TokenNumber,
		Lexeme:   l.lexeme(),
		Position: posicion,
	}
}

func (l *Lexer) stringToken(posicion Position) (Token, error) {
	for !l.isAtEnd() && l.peek() != '\'' {
		l.advance()
	}

	if l.isAtEnd() {
		return Token{}, &SyntaxError{
			Message:  "cadena de texto sin cerrar",
			Position: posicion,
		}
	}

	l.advance()

	return Token{
		Type:     TokenString,
		Lexeme:   l.lexeme(),
		Position: posicion,
	}, nil
}

func (l *Lexer) token(tipo TokenType, posicion Position) Token {
	return Token{
		Type:     tipo,
		Lexeme:   l.lexeme(),
		Position: posicion,
	}
}

func (l *Lexer) advance() rune {
	caracter := l.input[l.actual]
	l.actual++

	if caracter == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	return caracter
}

func (l *Lexer) match(esperado rune) bool {
	if l.isAtEnd() || l.peek() != esperado {
		return false
	}

	l.advance()
	return true
}

func (l *Lexer) peek() rune {
	return l.input[l.actual]
}

func (l *Lexer) peekNextIsDigit() bool {
	if l.actual+1 >= len(l.input) {
		return false
	}

	return unicode.IsDigit(l.input[l.actual+1])
}

func (l *Lexer) isAtEnd() bool {
	return l.actual >= len(l.input)
}

func (l *Lexer) lexeme() string {
	return string(l.input[l.start:l.actual])
}

func (l *Lexer) position() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.actual,
	}
}

func isIdentifierStart(caracter rune) bool {
	return unicode.IsLetter(caracter) || caracter == '_'
}

func isIdentifierPart(caracter rune) bool {
	return isIdentifierStart(caracter) || unicode.IsDigit(caracter)
}

func keywordTokenType(lexema string) TokenType {
	switch strings.ToUpper(lexema) {
	case "SELECT":
		return TokenSelect
	case "FROM":
		return TokenFrom
	case "WHERE":
		return TokenWhere
<<<<<<< HEAD
	case "INNER":
		return TokenInner
	case "JOIN":
		return TokenJoin
	case "ON":
		return TokenOn
=======
>>>>>>> af2c9a5137fac5ac5ffaed2e81ebc59fd20fca5a
	case "AND":
		return TokenAnd
	case "OR":
		return TokenOr
	case "TRUE":
		return TokenTrue
	case "FALSE":
		return TokenFalse
	case "NULL":
		return TokenNull
	case "ORDER":
		return TokenOrder
	case "BY":
		return TokenBy
	case "ASC":
		return TokenAsc
	case "DESC":
		return TokenDesc
	case "LIMIT":
		return TokenLimit
	case "GROUP":
		return TokenGroup
	case "HAVING":
		return TokenHaving
	case "COUNT":
		return TokenCount
	case "SUM":
		return TokenSum
	case "AVG":
		return TokenAvg
	case "MIN":
		return TokenMin
	case "MAX":
		return TokenMax
	case "IS":
		return TokenIs
	case "NOT":
		return TokenNot
	default:
		return TokenIdentifier
	}
}
