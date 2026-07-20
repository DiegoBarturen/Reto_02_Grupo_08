package main

import (
	"fmt"
	"io"
	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/executor"
	"reto_02_grupo_08/pkg/parser"
)

func main() {
	fmt.Println("Iniciando Motor SQL - Prueba de Evaluador Lógico...")

	// 1. Tabla simulada con datos básicos
	tablaPrueba := &almacenamiento.Tabla{
		Nombre: "amigos",
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoEntero, Dato: int64(1)},
				{Tipo: almacenamiento.TipoTexto, Dato: "Diego"},
			}},
			{Datos: []almacenamiento.Valor{
				{Tipo: almacenamiento.TipoEntero, Dato: int64(2)},
				{Tipo: almacenamiento.TipoTexto, Dato: "Abril"},
			}},
		},
	}

	// 2. Simulamos un AST con una operación lógica de tipo AND: true AND true
	astSimulado := &parser.SelectStmt{
		Columns: []parser.SelectColumn{{IsAsterisk: true}},
		Table:   "amigos",
		Where: &parser.BinaryExpr{
			Left:     &parser.Literal{Value: true},
			Operator: "AND",
			Right:    &parser.Literal{Value: true},
		},
	}

	// 3. Ensamblamos el pipeline completo
	var plan executor.Operator
	plan = executor.NuevoScanOperator(tablaPrueba)
	plan = executor.NuevoFilterOperator(plan, astSimulado.Where)
	plan = executor.NuevoProjectOperator(plan, astSimulado.Columns)

	defer plan.Close()

	// 4. Ejecución del árbol
	for {
		fila, err := plan.Next()
		if err == io.EOF {
			fmt.Println("--- Fin de la consulta de prueba ---")
			break
		}
		if err != nil {
			fmt.Printf("Error de ejecución: %v\n", err)
			break
		}
		fmt.Printf("Fila aprobada por el filtro lógico: %v\n", fila.Datos)
	}
}
