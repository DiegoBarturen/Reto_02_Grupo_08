package executor

import (
	"fmt"
	"io"
	"sort"

	"reto_02_grupo_08/pkg/almacenamiento"
	"reto_02_grupo_08/pkg/parser"
)

// OrdenColumna define un criterio de ordenamiento: índice de columna y dirección.
type OrdenColumna struct {
	Indice      int
	Descendente bool
}

// SortOperator implementa ORDER BY en el modelo Volcano.
// Materializa todas las filas del hijo, las ordena y las devuelve una a una.
// Los NULLs se ubican al final en ambas direcciones (NULLS LAST).
type SortOperator struct {
	hijo    Operator
	ordenes []OrdenColumna
	esquema almacenamiento.Esquema
	filas   []Row
	cursor  int
	cargado bool
}

// NuevoSortOperator construye un SortOperator resolviendo los nombres de columna
// del ORDER BY contra el esquema del operador hijo.
func NuevoSortOperator(hijo Operator, criterios []parser.OrderByItem) (*SortOperator, error) {
	esquema := hijo.Schema()
	ordenes := make([]OrdenColumna, len(criterios))

	for i, item := range criterios {
		idx, err := indiceDeLaColumna(esquema, item.Column)
		if err != nil {
			return nil, fmt.Errorf("ORDER BY: %w", err)
		}
		ordenes[i] = OrdenColumna{
			Indice:      idx,
			Descendente: item.Desc,
		}
	}

	return &SortOperator{
		hijo:    hijo,
		ordenes: ordenes,
		esquema: esquema,
	}, nil
}

func (s *SortOperator) Next() (Row, error) {
	if !s.cargado {
		if err := s.materializar(); err != nil {
			return Row{}, err
		}
		s.cargado = true
	}

	if s.cursor >= len(s.filas) {
		return Row{}, io.EOF
	}

	fila := s.filas[s.cursor]
	s.cursor++
	return fila, nil
}

// materializar carga todas las filas del hijo y las ordena en memoria.
func (s *SortOperator) materializar() error {
	s.filas = make([]Row, 0)

	for {
		fila, err := s.hijo.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		s.filas = append(s.filas, fila)
	}

	sort.SliceStable(s.filas, func(i, j int) bool {
		for _, orden := range s.ordenes {
			a := s.filas[i].Datos[orden.Indice]
			b := s.filas[j].Datos[orden.Indice]

			cmp, comparable := CompararValores(a, b)

			if !comparable {
				// NULLs siempre al final (NULLS LAST).
				if a.EsNulo && !b.EsNulo {
					return false
				}
				if !a.EsNulo && b.EsNulo {
					return true
				}
				continue // ambos son NULL: empate, probar siguiente criterio
			}

			if cmp == 0 {
				continue // empate: probar siguiente criterio
			}

			if orden.Descendente {
				return cmp > 0
			}
			return cmp < 0
		}
		return false
	})

	return nil
}

// Schema delega al hijo: el orden no cambia el esquema.
func (s *SortOperator) Schema() almacenamiento.Esquema {
	return s.esquema
}

func (s *SortOperator) Close() error {
	s.filas = nil
	s.cursor = 0
	s.cargado = false
	return s.hijo.Close()
}
