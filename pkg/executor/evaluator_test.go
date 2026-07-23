package executor

import (
	"io"
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// tablaDeEmpleados construye una tabla de prueba con datos representativos.
func tablaDeEmpleados() *almacenamiento.Tabla {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "id", Tipo: almacenamiento.TipoEntero},
			{Nombre: "nombre", Tipo: almacenamiento.TipoTexto},
			{Nombre: "salario", Tipo: almacenamiento.TipoDecimal},
		},
	}

	return &almacenamiento.Tabla{
		Nombre:  "empleados",
		Esquema: esquema,
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoEntero, Dato: int64(1)},
				{Tipo: almacenamiento.TipoTexto, Dato: "Ana"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 3000.0},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoEntero, Dato: int64(2)},
				{Tipo: almacenamiento.TipoTexto, Dato: "Carlos"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 5000.0},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoEntero, Dato: int64(3)},
				{Tipo: almacenamiento.TipoTexto, Dato: "María"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 4000.0},
			}},
		},
	}
}

// recolectarFilas drena un operador y retorna todas las filas.
func recolectarFilas(t *testing.T, op Operator) []Row {
	t.Helper()
	var filas []Row
	for {
		fila, err := op.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() devolvió error inesperado: %v", err)
		}
		filas = append(filas, fila)
	}
	return filas
}

// TestEvaluarLiteralYNull verifica la semántica NULL: comparaciones con NULL → false.
func TestEvaluarLiteralYNull(t *testing.T) {
	pruebas := []struct {
		nombre   string
		expr     parser.Expr
		esperado bool
	}{
		{
			nombre:   "literal true",
			expr:     &parser.Literal{Value: true},
			esperado: true, // true es true
		},
		{
			nombre: "NULL = 1 es false",
			expr: &parser.BinaryExpr{
				Left:     &parser.Literal{Value: nil},
				Operator: "=",
				Right:    &parser.Literal{Value: int64(1)},
			},
			esperado: false,
		},
		{
			nombre: "1 = NULL es false",
			expr: &parser.BinaryExpr{
				Left:     &parser.Literal{Value: int64(1)},
				Operator: "=",
				Right:    &parser.Literal{Value: nil},
			},
			esperado: false,
		},
	}

	esquema := almacenamiento.Esquema{}
	fila := Row{}

	for _, p := range pruebas {
		t.Run(p.nombre, func(t *testing.T) {
			resultado, err := Evaluar(fila, esquema, p.expr)
			if err != nil {
				t.Fatalf("Evaluar devolvió error inesperado: %v", err)
			}
			if b, ok := resultado.(bool); ok && b != p.esperado {
				t.Errorf("se esperaba %v y se obtuvo %v", p.esperado, resultado)
			}
		})
	}
}

// TestEvaluarComparacionesNumericas verifica operadores <, >, <=, >= sobre números.
func TestEvaluarComparacionesNumericas(t *testing.T) {
	pruebas := []struct {
		nombre   string
		izq      any
		op       string
		der      any
		esperado bool
	}{
		{"5 > 3", int64(5), ">", int64(3), true},
		{"3 > 5", int64(3), ">", int64(5), false},
		{"5 < 3", int64(5), "<", int64(3), false},
		{"3 < 5", int64(3), "<", int64(5), true},
		{"5 >= 5", int64(5), ">=", int64(5), true},
		{"4 >= 5", int64(4), ">=", int64(5), false},
		{"5 <= 5", int64(5), "<=", int64(5), true},
		{"6 <= 5", int64(6), "<=", int64(5), false},
		{"3.14 > 3", 3.14, ">", int64(3), true},
		{"abc <> def", "abc", "<>", "def", true},
		{"abc = abc", "abc", "=", "abc", true},
	}

	esquema := almacenamiento.Esquema{}
	fila := Row{}

	for _, p := range pruebas {
		t.Run(p.nombre, func(t *testing.T) {
			expr := &parser.BinaryExpr{
				Left:     &parser.Literal{Value: p.izq},
				Operator: p.op,
				Right:    &parser.Literal{Value: p.der},
			}
			resultado, err := Evaluar(fila, esquema, expr)
			if err != nil {
				t.Fatalf("Evaluar devolvió error: %v", err)
			}
			b, ok := resultado.(bool)
			if !ok {
				t.Fatalf("se esperaba bool y se obtuvo %T", resultado)
			}
			if b != p.esperado {
				t.Errorf("se esperaba %v y se obtuvo %v", p.esperado, b)
			}
		})
	}
}

// TestEvaluarIdentificadorDeColumna verifica que los identificadores resuelvan contra el esquema.
func TestEvaluarIdentificadorDeColumna(t *testing.T) {
	tabla := tablaDeEmpleados()
	fila := tabla.Filas[0] // Ana, salario=3000

	expr := &parser.BinaryExpr{
		Left:     &parser.Identifier{Name: "salario"},
		Operator: ">",
		Right:    &parser.Literal{Value: float64(2000)},
	}

	resultado, err := Evaluar(fila, tabla.Esquema, expr)
	if err != nil {
		t.Fatalf("Evaluar devolvió error: %v", err)
	}
	if b, ok := resultado.(bool); !ok || !b {
		t.Errorf("se esperaba true (salario 3000 > 2000) y se obtuvo %v", resultado)
	}
}

// TestEvaluarColumnaInexistente verifica el error para columnas que no existen.
func TestEvaluarColumnaInexistente(t *testing.T) {
	tabla := tablaDeEmpleados()
	fila := tabla.Filas[0]
	expr := &parser.Identifier{Name: "no_existe"}

	_, err := Evaluar(fila, tabla.Esquema, expr)
	if err == nil {
		t.Fatal("se esperaba error para columna inexistente")
	}
}

// TestProjectColumnaEspecifica verifica que Project filtra columnas correctamente.
func TestProjectColumnaEspecifica(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)

	// SELECT nombre, salario
	columnas := []parser.SelectColumn{
		{Name: "nombre"},
		{Name: "salario"},
	}
	proj, err := NuevoProjectOperator(scan, columnas)
	if err != nil {
		t.Fatalf("NuevoProjectOperator devolvió error: %v", err)
	}
	defer proj.Close()

	if len(proj.Schema().Columnas) != 2 {
		t.Fatalf("se esperaban 2 columnas en el esquema de salida y se obtuvieron %d",
			len(proj.Schema().Columnas))
	}
	if proj.Schema().Columnas[0].Nombre != "nombre" {
		t.Errorf("se esperaba columna 'nombre' y se obtuvo '%s'", proj.Schema().Columnas[0].Nombre)
	}

	filas := recolectarFilas(t, proj)
	if len(filas) != 3 {
		t.Fatalf("se esperaban 3 filas y se obtuvieron %d", len(filas))
	}
	// La primera fila debe tener 2 valores
	if len(filas[0].Datos) != 2 {
		t.Errorf("se esperaban 2 valores en la fila y se obtuvieron %d", len(filas[0].Datos))
	}
}
