package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/catalogo"
	"reto_02_grupo_08/pkg/executor"
	"reto_02_grupo_08/pkg/parser"
)

func main() {
	cat := catalogo.Nuevo()

	// Cargar CSVs desde la carpeta data/ (junto al binario o al directorio de trabajo).
	cargarCSVsDesdeDirectorio(cat, "data")

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Motor SQL en Memoria — Grupo 08        ║")
	fmt.Println("╠══════════════════════════════════════════╣")

	tablas := cat.ListarTablas()
	if len(tablas) > 0 {
		fmt.Printf("║  Tablas disponibles: %-20s║\n", strings.Join(tablas, ", "))
	} else {
		fmt.Println("║  Sin tablas. Coloca CSVs en data/        ║")
	}
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Comandos especiales: 'tablas' para listar, 'salir' para terminar.")
	fmt.Println()

	lector := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("sql> ")
		if !lector.Scan() {
			break
		}

		entrada := strings.TrimSpace(lector.Text())

		if entrada == "" {
			continue
		}

		switch strings.ToLower(entrada) {
		case "salir", "exit", "quit":
			fmt.Println("¡Hasta luego!")
			return
		case "tablas", "tables":
			fmt.Printf("Tablas: %s\n\n", strings.Join(cat.ListarTablas(), ", "))
			continue
		}

		ejecutarConsulta(entrada, cat)
		fmt.Println()
	}
}

// cargarCSVsDesdeDirectorio carga todos los archivos .csv del directorio indicado.
func cargarCSVsDesdeDirectorio(cat *catalogo.Catalogo, directorio string) {
	entradas, err := os.ReadDir(directorio)
	if err != nil {
		return // el directorio no existe: no es un error fatal
	}

	for _, entrada := range entradas {
		if entrada.IsDir() {
			continue
		}

		nombre := entrada.Name()
		if !strings.HasSuffix(strings.ToLower(nombre), ".csv") {
			continue
		}

		nombreTabla := strings.TrimSuffix(nombre, filepath.Ext(nombre))
		ruta := filepath.Join(directorio, nombre)

		tabla, err := almacenamiento.CargarCSVConInferencia(nombreTabla, ruta)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error cargando %s: %v\n", ruta, err)
			continue
		}

		if err := cat.RegistrarTabla(tabla); err != nil {
			fmt.Fprintf(os.Stderr, "✗ Error registrando '%s': %v\n", nombreTabla, err)
			continue
		}

		fmt.Printf("✓ Tabla '%s' cargada (%d filas, %d columnas)\n",
			nombreTabla, len(tabla.Filas), len(tabla.Esquema.Columnas))
	}
}

// ejecutarConsulta parsea y ejecuta una consulta SQL, imprimiendo los resultados.
func ejecutarConsulta(sql string, cat *catalogo.Catalogo) {
	stmt, err := parser.Parse(sql)
	if err != nil {
		fmt.Printf("Error de sintaxis: %v\n", err)
		return
	}

	if stmt.Join != nil {
		nestedJoin, err := construirJoinBruto(stmt, cat, "nested")
		if err != nil {
			fmt.Printf("Error al construir el plan: %v\n", err)
			return
		}
		defer nestedJoin.Close()

		hashJoin, err := construirJoinBruto(stmt, cat, "hash")
		if err != nil {
			fmt.Printf("Error al construir el plan: %v\n", err)
			return
		}
		defer hashJoin.Close()

		fmt.Printf("JOIN nested-loop: %s\n", nestedJoin.Duracion())
		fmt.Printf("JOIN hash: %s\n", hashJoin.Duracion())

		nestedPlan, err := construirPlanConAlgoritmo(stmt, cat, "nested")
		if err != nil {
			fmt.Printf("Error al construir el plan: %v\n", err)
			return
		}
		defer nestedPlan.Close()

		imprimirResultados(nestedPlan)
		return
	}

	plan, err := construirPlanConAlgoritmo(stmt, cat, "nested")
	if err != nil {
		fmt.Printf("Error al construir el plan: %v\n", err)
		return
	}
	defer plan.Close()

	imprimirResultados(plan)
}

// construirPlanConAlgoritmo ensambla el árbol de operadores Volcano a partir del AST.
// Orden: Scan → Join(ON) → Filter(WHERE) → Aggregate(GROUP BY) o Project(SELECT) → Sort(ORDER BY) → Limit
func construirJoinBruto(stmt *parser.SelectStmt, cat *catalogo.Catalogo, algoritmo string) (*executor.JoinOperator, error) {
	leftTabla, err := cat.ObtenerTabla(stmt.Table)
	if err != nil {
		return nil, err
	}

	if stmt.Join == nil {
		return nil, fmt.Errorf("la consulta no contiene JOIN")
	}

	rightTabla, err := cat.ObtenerTabla(stmt.Join.Table)
	if err != nil {
		return nil, err
	}

	joinPlan, err := executor.NuevoJoinOperator(executor.NuevoScanOperator(leftTabla), stmt.Table, rightTabla, stmt.Join.On, algoritmo)
	if err != nil {
		return nil, err
	}

	return joinPlan, nil
}

func construirPlanConAlgoritmo(stmt *parser.SelectStmt, cat *catalogo.Catalogo, algoritmo string) (executor.Operator, error) {
	leftTabla, err := cat.ObtenerTabla(stmt.Table)
	if err != nil {
		return nil, err
	}

	var plan executor.Operator = executor.NuevoScanOperator(leftTabla)

	if stmt.Join != nil {
		rightTabla, err := cat.ObtenerTabla(stmt.Join.Table)
		if err != nil {
			return nil, err
		}
		joinPlan, err := executor.NuevoJoinOperator(plan, stmt.Table, rightTabla, stmt.Join.On, algoritmo)
		if err != nil {
			return nil, err
		}
		plan = joinPlan
	}

	// WHERE
	if stmt.Where != nil {
		plan = executor.NuevoFilterOperator(plan, stmt.Where)
	}

	// GROUP BY / Agregación
	if len(stmt.GroupBy) > 0 || tieneAgregados(stmt.Columns) {
		aggPlan, err := executor.NuevoAggregateOperator(plan, stmt.GroupBy, stmt.Columns)
		if err != nil {
			return nil, err
		}
		plan = aggPlan
	} else {
		// Proyección simple de columnas
		projPlan, err := executor.NuevoProjectOperator(plan, stmt.Columns)
		if err != nil {
			return nil, err
		}
		plan = projPlan
	}

	// ORDER BY
	if len(stmt.OrderBy) > 0 {
		sortPlan, err := executor.NuevoSortOperator(plan, stmt.OrderBy)
		if err != nil {
			return nil, err
		}
		plan = sortPlan
	}

	// LIMIT
	if stmt.Limit != nil {
		plan = executor.NuevoLimitOperator(plan, int(*stmt.Limit))
	}

	return plan, nil
}

// tieneAgregados retorna true si alguna columna del SELECT es una función de agregación.
func tieneAgregados(columnas []parser.SelectColumn) bool {
	for _, col := range columnas {
		if col.Agg != nil {
			return true
		}
	}
	return false
}

// imprimirResultados muestra las filas en formato de tabla ASCII.
func imprimirResultados(plan executor.Operator) {
	esquema := plan.Schema()
	filas := make([]executor.Row, 0)

	for {
		fila, err := plan.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error de ejecución: %v\n", err)
			return
		}
		filas = append(filas, fila)
	}

	numCols := len(esquema.Columnas)

	// Calcular anchos de columna (mínimo: ancho del nombre).
	anchos := make([]int, numCols)
	for i, col := range esquema.Columnas {
		anchos[i] = len(col.Nombre)
	}

	// Convertir valores a texto y ajustar anchos.
	datos := make([][]string, len(filas))
	for i, fila := range filas {
		datos[i] = make([]string, numCols)
		for j := 0; j < numCols; j++ {
			if j >= len(fila.Datos) {
				datos[i][j] = ""
				continue
			}
			v := fila.Datos[j]
			if v.EsNulo {
				datos[i][j] = "NULL"
			} else {
				datos[i][j] = fmt.Sprint(v.Dato)
			}
			if len(datos[i][j]) > anchos[j] {
				anchos[j] = len(datos[i][j])
			}
		}
	}

	// Construir separador.
	sep := "+"
	for _, a := range anchos {
		sep += strings.Repeat("-", a+2) + "+"
	}

	// Encabezado.
	fmt.Println(sep)
	header := "|"
	for i, col := range esquema.Columnas {
		header += fmt.Sprintf(" %-*s |", anchos[i], col.Nombre)
	}
	fmt.Println(header)
	fmt.Println(sep)

	// Filas de datos.
	for _, fila := range datos {
		linea := "|"
		for i, celda := range fila {
			linea += fmt.Sprintf(" %-*s |", anchos[i], celda)
		}
		fmt.Println(linea)
	}

	fmt.Println(sep)

	if len(filas) == 1 {
		fmt.Printf("(1 fila)\n")
	} else {
		fmt.Printf("(%d filas)\n", len(filas))
	}
}
