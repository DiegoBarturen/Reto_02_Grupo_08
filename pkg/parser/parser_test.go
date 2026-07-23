package parser

import "testing"

func TestParseSelectConAsterisco(t *testing.T) {
	stmt, err := Parse("SELECT * FROM estudiantes")
	if err != nil {
		t.Fatalf("Parse devolvio un error inesperado: %v", err)
	}

	if stmt.Table != "estudiantes" {
		t.Errorf(
			"se esperaba la tabla %q y se obtuvo %q",
			"estudiantes",
			stmt.Table,
		)
	}

	if len(stmt.Columns) != 1 || !stmt.Columns[0].IsAsterisk {
		t.Fatalf("se esperaba una proyeccion con * y se obtuvo %#v", stmt.Columns)
	}

	if stmt.Where != nil {
		t.Errorf("no se esperaba WHERE y se obtuvo %#v", stmt.Where)
	}
}

func TestParseSelectConColumnasYWhere(t *testing.T) {
	stmt, err := Parse("SELECT id, nombre FROM estudiantes WHERE edad >= 18 AND activo = true")
	if err != nil {
		t.Fatalf("Parse devolvio un error inesperado: %v", err)
	}

	if len(stmt.Columns) != 2 {
		t.Fatalf(
			"se esperaban 2 columnas y se obtuvieron %d",
			len(stmt.Columns),
		)
	}

	if stmt.Columns[0].Name != "id" || stmt.Columns[1].Name != "nombre" {
		t.Errorf("columnas incorrectas: %#v", stmt.Columns)
	}

	expresion, correcto := stmt.Where.(*BinaryExpr)
	if !correcto {
		t.Fatalf("se esperaba BinaryExpr en WHERE y se obtuvo %T", stmt.Where)
	}

	if expresion.Operator != "AND" {
		t.Errorf("se esperaba operador AND y se obtuvo %q", expresion.Operator)
	}
}

func TestParsePrecedenciaAndAntesQueOr(t *testing.T) {
	stmt, err := Parse("SELECT * FROM estudiantes WHERE id = 1 OR activo = true AND promedio >= 15")
	if err != nil {
		t.Fatalf("Parse devolvio un error inesperado: %v", err)
	}

	raiz, correcto := stmt.Where.(*BinaryExpr)
	if !correcto {
		t.Fatalf("se esperaba BinaryExpr y se obtuvo %T", stmt.Where)
	}

	if raiz.Operator != "OR" {
		t.Fatalf("se esperaba OR en la raiz y se obtuvo %q", raiz.Operator)
	}

	derecha, correcto := raiz.Right.(*BinaryExpr)
	if !correcto {
		t.Fatalf("se esperaba BinaryExpr al lado derecho y se obtuvo %T", raiz.Right)
	}

	if derecha.Operator != "AND" {
		t.Errorf("se esperaba AND al lado derecho y se obtuvo %q", derecha.Operator)
	}
}

func TestParseParentesisAlteranPrecedencia(t *testing.T) {
	stmt, err := Parse("SELECT * FROM estudiantes WHERE (id = 1 OR activo = true) AND promedio >= 15")
	if err != nil {
		t.Fatalf("Parse devolvio un error inesperado: %v", err)
	}

	raiz, correcto := stmt.Where.(*BinaryExpr)
	if !correcto {
		t.Fatalf("se esperaba BinaryExpr y se obtuvo %T", stmt.Where)
	}

	if raiz.Operator != "AND" {
		t.Fatalf("se esperaba AND en la raiz y se obtuvo %q", raiz.Operator)
	}

	izquierda, correcto := raiz.Left.(*BinaryExpr)
	if !correcto {
		t.Fatalf("se esperaba BinaryExpr al lado izquierdo y se obtuvo %T", raiz.Left)
	}

	if izquierda.Operator != "OR" {
		t.Errorf("se esperaba OR al lado izquierdo y se obtuvo %q", izquierda.Operator)
	}
}

func TestParseInnerJoinOn(t *testing.T) {
	stmt, err := Parse("SELECT empleados.id, ventas.producto FROM empleados INNER JOIN ventas ON empleados.id = ventas.id")
	if err != nil {
		t.Fatalf("Parse devolvio un error inesperado: %v", err)
	}

	if stmt.Join == nil {
		t.Fatal("se esperaba una cláusula JOIN en el AST")
	}
	if stmt.Join.Table != "ventas" {
		t.Fatalf("se esperaba unir la tabla %q y se obtuvo %q", "ventas", stmt.Join.Table)
	}
	if stmt.Join.On == nil {
		t.Fatal("se esperaba una condición ON en la cláusula JOIN")
	}
}

func TestParseOperadoresComparacion(t *testing.T) {
	consultas := []string{
		"SELECT * FROM t WHERE a = 1",
		"SELECT * FROM t WHERE a <> 1",
		"SELECT * FROM t WHERE a < 1",
		"SELECT * FROM t WHERE a > 1",
		"SELECT * FROM t WHERE a <= 1",
		"SELECT * FROM t WHERE a >= 1",
	}

	for _, consulta := range consultas {
		t.Run(consulta, func(t *testing.T) {
			if _, err := Parse(consulta); err != nil {
				t.Fatalf("Parse devolvio un error inesperado: %v", err)
			}
		})
	}
}

func TestParseLiterales(t *testing.T) {
	consultas := []string{
		"SELECT * FROM t WHERE nombre = 'Ana'",
		"SELECT * FROM t WHERE activo = FALSE",
		"SELECT * FROM t WHERE promedio = 17.5",
		"SELECT * FROM t WHERE observacion = NULL",
	}

	for _, consulta := range consultas {
		t.Run(consulta, func(t *testing.T) {
			if _, err := Parse(consulta); err != nil {
				t.Fatalf("Parse devolvio un error inesperado: %v", err)
			}
		})
	}
}

func TestParseErroresSintaxis(t *testing.T) {
	pruebas := []struct {
		nombre          string
		consulta        string
		columnaEsperada int
	}{
		{
			nombre:          "sin SELECT",
			consulta:        "id FROM estudiantes",
			columnaEsperada: 1,
		},
		{
			nombre:          "coma sin columna",
			consulta:        "SELECT id, FROM estudiantes",
			columnaEsperada: 12,
		},
		{
			nombre:          "sin FROM",
			consulta:        "SELECT id estudiantes",
			columnaEsperada: 11,
		},
		{
			nombre:          "sin tabla",
			consulta:        "SELECT id FROM",
			columnaEsperada: 15,
		},
		{
			nombre:          "parentesis sin cerrar",
			consulta:        "SELECT * FROM estudiantes WHERE (edad >= 18",
			columnaEsperada: 44,
		},
		{
			nombre:          "tokens sobrantes",
			consulta:        "SELECT * FROM estudiantes WHERE edad >= 18 nombre",
			columnaEsperada: 44,
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			_, err := Parse(prueba.consulta)
			if err == nil {
				t.Fatal("se esperaba un error de sintaxis")
			}

			errorSintaxis, correcto := err.(*SyntaxError)
			if !correcto {
				t.Fatalf("se esperaba SyntaxError y se obtuvo %T", err)
			}

			if errorSintaxis.Position.Column != prueba.columnaEsperada {
				t.Errorf(
					"se esperaba columna %d y se obtuvo %d: %v",
					prueba.columnaEsperada,
					errorSintaxis.Position.Column,
					err,
				)
			}
		})
	}
}
