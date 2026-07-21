package executor

import "reto_02_grupo_08/pkg/almacenamiento"

type Fila = almacenamiento.Fila

type Row = Fila

// Operator es la interfaz de iterador del modelo Volcano.
// Cada operador produce filas de una en una con Next(),
// libera recursos con Close() y expone su esquema de salida con Schema().
// Agregar un operador nuevo no requiere modificar los existentes.
type Operator interface {
	Next() (Row, error)
	Close() error
	Schema() almacenamiento.Esquema
}
