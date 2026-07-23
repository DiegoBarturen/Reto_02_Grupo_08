package parser

import "testing"

func TestParseSelectConJoinYColumnasCalificadas(t *testing.T) {
	stmt, err := Parse("SELECT empleados.id, ventas.producto FROM empleados INNER JOIN ventas ON empleados.id = ventas.id")
	if err != nil {
		t.Fatalf("Parse devolvió error: %v", err)
	}

	if stmt.Table != "empleados" {
		t.Fatalf("se esperaba tabla 'empleados' y se obtuvo %q", stmt.Table)
	}
	if stmt.Join == nil {
		t.Fatal("se esperaba JOIN en el AST")
	}
	if stmt.Join.Table != "ventas" {
		t.Fatalf("se esperaba tabla derecha 'ventas' y se obtuvo %q", stmt.Join.Table)
	}
	if len(stmt.Columns) != 2 {
		t.Fatalf("se esperaban 2 columnas seleccionadas y se obtuvieron %d", len(stmt.Columns))
	}
}
