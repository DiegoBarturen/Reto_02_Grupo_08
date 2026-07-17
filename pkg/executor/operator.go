package executor

import "reto_02_grupo_08/pkg/storage"

type Row = storage.Row

type Operator interface {
	Next() (Row, error)
	Close() error
}
