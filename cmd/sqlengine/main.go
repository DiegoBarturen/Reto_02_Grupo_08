package main

import (
	"fmt"
	"os"
	"strings"

	"reto_02_grupo_08/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Motor SQL en Memoria")
		fmt.Println("Uso:")
		fmt.Println(`  go run ./cmd/sqlengine "SELECT * FROM estudiantes WHERE edad >= 18"`)
		return
	}

	consulta := strings.Join(os.Args[1:], " ")

	tokens, err := parser.Lex(consulta)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stmt, err := parser.Parse(consulta)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Consulta:")
	fmt.Println(consulta)
	fmt.Println()
	fmt.Println("Tokens:")

	for _, token := range tokens {
		fmt.Printf(
			"  %-10s %-15q linea=%d columna=%d\n",
			token.Type,
			token.Lexeme,
			token.Position.Line,
			token.Position.Column,
		)
	}

	fmt.Println()
	fmt.Println("AST:")
	imprimirSelect(stmt)
}

func imprimirSelect(stmt *parser.SelectStmt) {
	fmt.Printf("  SelectStmt\n")
	fmt.Printf("    Tabla: %s\n", stmt.Table)
	fmt.Printf("    Columnas:\n")

	for _, columna := range stmt.Columns {
		fmt.Printf("      - %s\n", columna.String())
	}

	if stmt.Where == nil {
		fmt.Printf("    Where: <sin condicion>\n")
		return
	}

	fmt.Printf("    Where:\n")
	imprimirExpr(stmt.Where, "      ")
}

func imprimirExpr(expr parser.Expr, indentacion string) {
	switch expresion := expr.(type) {
	case *parser.BinaryExpr:
		fmt.Printf("%sBinaryExpr operador=%s\n", indentacion, expresion.Operator)
		fmt.Printf("%s  Izquierda:\n", indentacion)
		imprimirExpr(expresion.Left, indentacion+"    ")
		fmt.Printf("%s  Derecha:\n", indentacion)
		imprimirExpr(expresion.Right, indentacion+"    ")
	case *parser.Identifier:
		fmt.Printf("%sIdentifier nombre=%s\n", indentacion, expresion.Name)
	case *parser.Literal:
		fmt.Printf("%sLiteral valor=%v raw=%q\n", indentacion, expresion.Value, expresion.Raw)
	default:
		fmt.Printf("%s%T\n", indentacion, expr)
	}
}
