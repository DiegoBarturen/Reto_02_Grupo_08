package executor

import (
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// TestSortAscendente verifica que ORDER BY col ASC ordena de menor a mayor.
func TestSortAscendente(t *testing.T) {
	tabla := tablaDeEmpleados() // Ana=3000, Carlos=5000, María=4000
	scan := NuevoScanOperator(tabla)

	sort, err := NuevoSortOperator(scan, []parser.OrderByItem{
		{Column: "salario", Desc: false},
	})
	if err != nil {
		t.Fatalf("NuevoSortOperator devolvió error: %v", err)
	}
	defer sort.Close()

	filas := recolectarFilas(t, sort)
	if len(filas) != 3 {
		t.Fatalf("se esperaban 3 filas y se obtuvieron %d", len(filas))
	}

	// Índice 2 = salario
	salarios := []float64{
		filas[0].Datos[2].Dato.(float64),
		filas[1].Datos[2].Dato.(float64),
		filas[2].Datos[2].Dato.(float64),
	}
	if salarios[0] > salarios[1] || salarios[1] > salarios[2] {
		t.Errorf("orden incorrecto: %v", salarios)
	}
}

// TestSortDescendente verifica que ORDER BY col DESC ordena de mayor a menor.
func TestSortDescendente(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)

	sort, err := NuevoSortOperator(scan, []parser.OrderByItem{
		{Column: "salario", Desc: true},
	})
	if err != nil {
		t.Fatalf("NuevoSortOperator devolvió error: %v", err)
	}
	defer sort.Close()

	filas := recolectarFilas(t, sort)

	salarios := []float64{
		filas[0].Datos[2].Dato.(float64),
		filas[1].Datos[2].Dato.(float64),
		filas[2].Datos[2].Dato.(float64),
	}
	if salarios[0] < salarios[1] || salarios[1] < salarios[2] {
		t.Errorf("orden descendente incorrecto: %v", salarios)
	}
}

// TestSortConNull verifica que los NULLs se ubican al final (NULLS LAST).
func TestSortConNull(t *testing.T) {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "nombre", Tipo: almacenamiento.TipoTexto},
			{Nombre: "puntos", Tipo: almacenamiento.TipoEntero},
		},
	}
	tabla := &almacenamiento.Tabla{
		Nombre:  "t",
		Esquema: esquema,
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "B"},
				{Tipo: almacenamiento.TipoEntero, EsNulo: true},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "A"},
				{Tipo: almacenamiento.TipoEntero, Dato: int64(10)},
			}},
		},
	}

	scan := NuevoScanOperator(tabla)
	srt, err := NuevoSortOperator(scan, []parser.OrderByItem{
		{Column: "puntos", Desc: false},
	})
	if err != nil {
		t.Fatalf("NuevoSortOperator devolvió error: %v", err)
	}
	defer srt.Close()

	filas := recolectarFilas(t, srt)

	// El NULL debe ir al final: A (10), B (NULL)
	primerNombre := filas[0].Datos[0].Dato.(string)
	if primerNombre != "A" {
		t.Errorf("el NULL debería estar al final: primer elemento es '%s', esperado 'A'", primerNombre)
	}
	if !filas[1].Datos[1].EsNulo {
		t.Errorf("el segundo elemento debería ser NULL")
	}
}

// TestSortTablaVacia verifica que ordenar una tabla vacía no produce error.
func TestSortTablaVacia(t *testing.T) {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "id", Tipo: almacenamiento.TipoEntero},
		},
	}
	tabla := &almacenamiento.Tabla{Nombre: "vacia", Esquema: esquema, Filas: nil}
	scan := NuevoScanOperator(tabla)

	srt, err := NuevoSortOperator(scan, []parser.OrderByItem{
		{Column: "id", Desc: false},
	})
	if err != nil {
		t.Fatalf("NuevoSortOperator devolvió error: %v", err)
	}
	defer srt.Close()

	filas := recolectarFilas(t, srt)
	if len(filas) != 0 {
		t.Errorf("se esperaban 0 filas y se obtuvieron %d", len(filas))
	}
}

// TestSortColumnaInexistente verifica que se retorna error para columnas inválidas.
func TestSortColumnaInexistente(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)

	_, err := NuevoSortOperator(scan, []parser.OrderByItem{
		{Column: "no_existe", Desc: false},
	})
	if err == nil {
		t.Fatal("se esperaba error para columna inexistente en ORDER BY")
	}
}
