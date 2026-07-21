package executor

import (
	"fmt"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// ProjectOperator implementa la cláusula SELECT del modelo Volcano.
// Filtra las columnas de cada fila para devolver solo las pedidas por la consulta.
type ProjectOperator struct {
	hijo          Operator
	columnas      []parser.SelectColumn
	esquemaSalida almacenamiento.Esquema
	indicesSal    []int // mapeo índice-de-salida → índice-de-entrada
}

// NuevoProjectOperator construye un ProjectOperator validando que todas las
// columnas solicitadas existan en el esquema del hijo.
func NuevoProjectOperator(hijo Operator, columnas []parser.SelectColumn) (*ProjectOperator, error) {
	esquemaEntrada := hijo.Schema()

	// SELECT * → devolver todas las columnas sin transformación.
	if len(columnas) == 1 && columnas[0].IsAsterisk {
		return &ProjectOperator{
			hijo:          hijo,
			columnas:      columnas,
			esquemaSalida: esquemaEntrada,
		}, nil
	}

	colsSalida := make([]almacenamiento.Columna, 0, len(columnas))
	indices := make([]int, 0, len(columnas))

	for _, col := range columnas {
		if col.IsAsterisk {
			return nil, fmt.Errorf("SELECT * no puede combinarse con columnas específicas")
		}

		// Las columnas de agregación se buscan por su nombre de cadena (e.g. "COUNT(*)").
		nombreBuscar := col.Name
		if col.Agg != nil {
			nombreBuscar = col.Agg.String()
		}

		idx, err := indiceDeLaColumna(esquemaEntrada, nombreBuscar)
		if err != nil {
			return nil, err
		}
		colsSalida = append(colsSalida, esquemaEntrada.Columnas[idx])
		indices = append(indices, idx)
	}

	return &ProjectOperator{
		hijo:          hijo,
		columnas:      columnas,
		esquemaSalida: almacenamiento.Esquema{Columnas: colsSalida},
		indicesSal:    indices,
	}, nil
}

func (p *ProjectOperator) Next() (Row, error) {
	fila, err := p.hijo.Next()
	if err != nil {
		return Row{}, err
	}

	// SELECT * → la fila pasa sin transformación.
	if len(p.columnas) == 1 && p.columnas[0].IsAsterisk {
		return fila, nil
	}

	valores := make([]almacenamiento.Valor, len(p.indicesSal))
	for i, idx := range p.indicesSal {
		valores[i] = fila.Datos[idx]
	}
	return Row{Datos: valores}, nil
}

// Schema devuelve el esquema de salida (solo las columnas seleccionadas).
func (p *ProjectOperator) Schema() almacenamiento.Esquema {
	return p.esquemaSalida
}

func (p *ProjectOperator) Close() error {
	return p.hijo.Close()
}
