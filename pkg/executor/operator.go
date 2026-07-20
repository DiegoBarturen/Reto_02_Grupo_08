package executor

import "reto_02_grupo_08/pkg/almacenamiento"

type Fila = almacenamiento.Fila

type Row = Fila

type Operator interface {
	Next() (Row, error)
	Close() error
}
