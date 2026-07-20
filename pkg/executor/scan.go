package executor

import (
	"io"
	"reto_02_grupo_08/pkg/almacenamiento"
)

type ScanOperator struct {
	tabla  *almacenamiento.Tabla
	cursor int
}

func NuevoScanOperator(tabla *almacenamiento.Tabla) *ScanOperator {
	return &ScanOperator{
		tabla:  tabla,
		cursor: 0,
	}
}

func (s *ScanOperator) Next() (Row, error) {

	if s.cursor >= len(s.tabla.Filas) {
		return Row{}, io.EOF
	}

	filaActual := s.tabla.Filas[s.cursor]

	s.cursor++

	return filaActual, nil
}

func (s *ScanOperator) Close() error {
	s.cursor = 0
	return nil
}
