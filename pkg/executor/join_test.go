package executor

import (
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

func TestJoinOperatorNestedYHashRetornanMismasFilas(t *testing.T) {
	left := &almacenamiento.Tabla{
		Nombre: "empleados",
		Esquema: almacenamiento.Esquema{
			Columnas: []almacenamiento.Columna{
				{Nombre: "id", Tipo: almacenamiento.TipoEntero},
				{Nombre: "nombre", Tipo: almacenamiento.TipoTexto},
			},
		},
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoEntero, Dato: int64(1)}, {Tipo: almacenamiento.TipoTexto, Dato: "Ana"}}},
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoEntero, Dato: int64(2)}, {Tipo: almacenamiento.TipoTexto, Dato: "Carlos"}}},
		},
	}

	right := &almacenamiento.Tabla{
		Nombre: "ventas",
		Esquema: almacenamiento.Esquema{
			Columnas: []almacenamiento.Columna{
				{Nombre: "id", Tipo: almacenamiento.TipoEntero},
				{Nombre: "producto", Tipo: almacenamiento.TipoTexto},
			},
		},
		Filas: []almacenamiento.Fila{
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoEntero, Dato: int64(1)}, {Tipo: almacenamiento.TipoTexto, Dato: "Laptop"}}},
			{Datos: []almacenamiento.Valor{{Tipo: almacenamiento.TipoEntero, Dato: int64(2)}, {Tipo: almacenamiento.TipoTexto, Dato: "Mouse"}}},
		},
	}

	condicion := &parser.BinaryExpr{
		Left:     &parser.Identifier{Name: "empleados.id"},
		Operator: "=",
		Right:    &parser.Identifier{Name: "ventas.id"},
	}

	nested, err := NuevoJoinOperator(NuevoScanOperator(left), left.Nombre, right, condicion, "nested")
	if err != nil {
		t.Fatalf("NuevoJoinOperator(nested) devolvió error: %v", err)
	}
	defer nested.Close()

	hash, err := NuevoJoinOperator(NuevoScanOperator(left), left.Nombre, right, condicion, "hash")
	if err != nil {
		t.Fatalf("NuevoJoinOperator(hash) devolvió error: %v", err)
	}
	defer hash.Close()

	nestedRows := recolectarFilas(t, nested)
	hashRows := recolectarFilas(t, hash)

	if len(nestedRows) != 2 {
		t.Fatalf("se esperaban 2 filas en nested-loop y se obtuvieron %d", len(nestedRows))
	}
	if len(hashRows) != 2 {
		t.Fatalf("se esperaban 2 filas en hash join y se obtuvieron %d", len(hashRows))
	}

	if len(nested.Schema().Columnas) != 4 {
		t.Fatalf("se esperaban 4 columnas en el esquema del JOIN y se obtuvieron %d", len(nested.Schema().Columnas))
	}
}
