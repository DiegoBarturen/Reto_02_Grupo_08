package parser

import "testing"

func TestLexReconoceJoinYColumnasCalificadas(t *testing.T) {
	tokens, err := Lex("SELECT empleados.id FROM empleados INNER JOIN ventas ON empleados.id = ventas.id")
	if err != nil {
		t.Fatalf("Lex devolvió error: %v", err)
	}

	seenJoin := false
	seenOn := false
	seenDot := false

	for _, tok := range tokens {
		switch tok.Type {
		case TokenJoin:
			seenJoin = true
		case TokenOn:
			seenOn = true
		case TokenDot:
			seenDot = true
		}
	}

	if !seenJoin {
		t.Fatal("se esperaba reconocer JOIN")
	}
	if !seenOn {
		t.Fatal("se esperaba reconocer ON")
	}
	if !seenDot {
		t.Fatal("se esperaba reconocer el punto para columnas calificadas")
	}
}
