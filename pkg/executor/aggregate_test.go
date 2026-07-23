package executor

import (
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// tablaConNulls construye una tabla con valores NULL para probar los agregados.
func tablaConNulls() *almacenamiento.Tabla {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "depto", Tipo: almacenamiento.TipoTexto},
			{Nombre: "salario", Tipo: almacenamiento.TipoDecimal},
		},
	}
	return &almacenamiento.Tabla{
		Nombre:  "empleados",
		Esquema: esquema,
		Filas: []almacenamiento.Fila{
			// Depto A: salarios 1000, 2000
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "A"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 1000.0},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "A"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 2000.0},
			}},
			// Depto B: salario 3000 y un NULL
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "B"},
				{Tipo: almacenamiento.TipoDecimal, Dato: 3000.0},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoTexto, Dato: "B"},
				{Tipo: almacenamiento.TipoDecimal, EsNulo: true},
			}},
		},
	}
}

// TestCountStar verifica COUNT(*) — cuenta todas las filas incluyendo NULLs.
func TestCountStar(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "COUNT", IsStar: true}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	if len(filas) != 1 {
		t.Fatalf("se esperaba 1 fila y se obtuvieron %d", len(filas))
	}

	count, ok := filas[0].Datos[0].Dato.(int64)
	if !ok {
		t.Fatalf("se esperaba int64 y se obtuvo %T", filas[0].Datos[0].Dato)
	}
	if count != 4 {
		t.Errorf("COUNT(*) se esperaba 4 y se obtuvo %d", count)
	}
}

// TestCountColumnaIgnoraNull verifica COUNT(col) — no cuenta valores NULL.
func TestCountColumnaIgnoraNull(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "COUNT", Column: "salario"}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	count := filas[0].Datos[0].Dato.(int64)
	if count != 3 { // 4 filas pero 1 NULL → 3
		t.Errorf("COUNT(salario) se esperaba 3 (ignorando NULL) y se obtuvo %d", count)
	}
}

// TestSumConNull verifica que SUM ignora los valores NULL.
func TestSumConNull(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "SUM", Column: "salario"}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	suma := filas[0].Datos[0].Dato.(float64)
	if suma != 6000.0 { // 1000 + 2000 + 3000 (el NULL se ignora)
		t.Errorf("SUM(salario) se esperaba 6000 y se obtuvo %v", suma)
	}
}

// TestAvgConNull verifica que AVG ignora los valores NULL en el denominador.
func TestAvgConNull(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "AVG", Column: "salario"}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	avg := filas[0].Datos[0].Dato.(float64)
	// 6000 / 3 = 2000 (el NULL no cuenta ni en suma ni en denominador)
	if avg != 2000.0 {
		t.Errorf("AVG(salario) se esperaba 2000 y se obtuvo %v", avg)
	}
}

// TestGroupByConAgregados verifica GROUP BY con COUNT(*) y SUM.
func TestGroupByConAgregados(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Name: "depto"},
		{Agg: &parser.AggFunc{Name: "COUNT", IsStar: true}},
		{Agg: &parser.AggFunc{Name: "SUM", Column: "salario"}},
	}

	agg, err := NuevoAggregateOperator(scan, []string{"depto"}, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	if len(filas) != 2 { // grupos A y B
		t.Fatalf("se esperaban 2 grupos y se obtuvieron %d", len(filas))
	}

	// Verificar esquema de salida: depto, COUNT(*), SUM(salario)
	esquema := agg.Schema()
	if len(esquema.Columnas) != 3 {
		t.Fatalf("se esperaban 3 columnas en la salida y se obtuvieron %d", len(esquema.Columnas))
	}
}

// TestMinMaxConNull verifica que MIN/MAX ignoran los NULL.
func TestMinMaxConNull(t *testing.T) {
	tabla := tablaConNulls()
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "MIN", Column: "salario"}},
		{Agg: &parser.AggFunc{Name: "MAX", Column: "salario"}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	min := filas[0].Datos[0].Dato.(float64)
	max := filas[0].Datos[1].Dato.(float64)

	if min != 1000.0 {
		t.Errorf("MIN(salario) se esperaba 1000 y se obtuvo %v", min)
	}
	if max != 3000.0 {
		t.Errorf("MAX(salario) se esperaba 3000 y se obtuvo %v", max)
	}
}

// TestAggregateTablaVacia verifica que COUNT(*) de una tabla vacía devuelve 0.
func TestAggregateTablaVacia(t *testing.T) {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "valor", Tipo: almacenamiento.TipoEntero},
		},
	}
	tabla := &almacenamiento.Tabla{Nombre: "vacia", Esquema: esquema, Filas: nil}
	scan := NuevoScanOperator(tabla)

	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "COUNT", IsStar: true}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	if len(filas) != 1 {
		t.Fatalf("se esperaba 1 fila (escalar) con tabla vacía y se obtuvieron %d", len(filas))
	}

	count := filas[0].Datos[0].Dato.(int64)
	if count != 0 {
		t.Errorf("COUNT(*) de tabla vacía se esperaba 0 y se obtuvo %d", count)
	}
}

// TestSumTodosNullDevuelveNull verifica que SUM de solo NULLs devuelve NULL.
func TestSumTodosNullDevuelveNull(t *testing.T) {
	esquema := almacenamiento.Esquema{
		Columnas: []almacenamiento.Columna{
			{Nombre: "valor", Tipo: almacenamiento.TipoDecimal},
		},
	}
	tabla := &almacenamiento.Tabla{
		Nombre:  "todos_null",
		Esquema: esquema,
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoDecimal, EsNulo: true}}},
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoDecimal, EsNulo: true}}},
		},
	}

	scan := NuevoScanOperator(tabla)
	columnasSel := []parser.SelectColumn{
		{Agg: &parser.AggFunc{Name: "SUM", Column: "valor"}},
	}

	agg, err := NuevoAggregateOperator(scan, nil, columnasSel)
	if err != nil {
		t.Fatalf("NuevoAggregateOperator devolvió error: %v", err)
	}
	defer agg.Close()

	filas := recolectarFilas(t, agg)
	if !filas[0].Datos[0].EsNulo {
		t.Errorf("SUM de solo NULLs debería devolver NULL, se obtuvo %v", filas[0].Datos[0].Dato)
	}
}
