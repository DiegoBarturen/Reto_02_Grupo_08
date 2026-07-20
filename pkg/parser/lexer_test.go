package parser

import (
	"reflect"
	"testing"
)

func TestLexTokensConsultaSelect(t *testing.T) {
	tokens, err := Lex("SELECT id, nombre FROM estudiantes WHERE edad >= 18 AND activo = true")
	if err != nil {
		t.Fatalf("Lex devolvio un error inesperado: %v", err)
	}

	tiposEsperados := []TokenType{
		TokenSelect,
		TokenIdentifier,
		TokenComma,
		TokenIdentifier,
		TokenFrom,
		TokenIdentifier,
		TokenWhere,
		TokenIdentifier,
		TokenGreatEqual,
		TokenNumber,
		TokenAnd,
		TokenIdentifier,
		TokenEqual,
		TokenTrue,
		TokenEOF,
	}

	tiposObtenidos := make([]TokenType, len(tokens))
	for indice, token := range tokens {
		tiposObtenidos[indice] = token.Type
	}

	if !reflect.DeepEqual(tiposObtenidos, tiposEsperados) {
		t.Errorf(
			"se esperaban los tokens %v y se obtuvo %v",
			tiposEsperados,
			tiposObtenidos,
		)
	}
}

func TestLexPosicionToken(t *testing.T) {
	tokens, err := Lex("SELECT *\nFROM estudiantes")
	if err != nil {
		t.Fatalf("Lex devolvio un error inesperado: %v", err)
	}

	from := tokens[2]

	if from.Type != TokenFrom {
		t.Fatalf("se esperaba FROM y se obtuvo %s", from.Type)
	}

	if from.Position.Line != 2 || from.Position.Column != 1 {
		t.Errorf(
			"se esperaba posicion linea 2, columna 1 y se obtuvo %s",
			from.Position.String(),
		)
	}
}

func TestLexErrorCaracterInvalido(t *testing.T) {
	_, err := Lex("SELECT @ FROM estudiantes")
	if err == nil {
		t.Fatal("se esperaba un error por caracter invalido")
	}

	errorSintaxis, correcto := err.(*SyntaxError)
	if !correcto {
		t.Fatalf("se esperaba SyntaxError y se obtuvo %T", err)
	}

	if errorSintaxis.Position.Column != 8 {
		t.Errorf(
			"se esperaba columna 8 y se obtuvo %d",
			errorSintaxis.Position.Column,
		)
	}
}

func TestLexErrorCadenaSinCerrar(t *testing.T) {
	_, err := Lex("SELECT * FROM estudiantes WHERE nombre = 'Ana")
	if err == nil {
		t.Fatal("se esperaba un error por cadena sin cerrar")
	}
}
