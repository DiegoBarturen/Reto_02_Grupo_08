package executor

import (
	"testing"
)

// TestLimitNormal verifica que LIMIT n emite exactamente n filas.
func TestLimitNormal(t *testing.T) {
	tabla := tablaDeEmpleados() // 3 filas
	scan := NuevoScanOperator(tabla)
	lim := NuevoLimitOperator(scan, 2)
	defer lim.Close()

	filas := recolectarFilas(t, lim)
	if len(filas) != 2 {
		t.Errorf("se esperaban 2 filas con LIMIT 2 y se obtuvieron %d", len(filas))
	}
}

// TestLimitMayorQueTotal verifica que LIMIT n > len(tabla) emite todas las filas.
func TestLimitMayorQueTotal(t *testing.T) {
	tabla := tablaDeEmpleados() // 3 filas
	scan := NuevoScanOperator(tabla)
	lim := NuevoLimitOperator(scan, 100)
	defer lim.Close()

	filas := recolectarFilas(t, lim)
	if len(filas) != 3 {
		t.Errorf("se esperaban 3 filas con LIMIT 100 y se obtuvieron %d", len(filas))
	}
}

// TestLimitCero verifica que LIMIT 0 no emite ninguna fila.
func TestLimitCero(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)
	lim := NuevoLimitOperator(scan, 0)
	defer lim.Close()

	filas := recolectarFilas(t, lim)
	if len(filas) != 0 {
		t.Errorf("se esperaban 0 filas con LIMIT 0 y se obtuvieron %d", len(filas))
	}
}

// TestLimitUno verifica que LIMIT 1 emite exactamente 1 fila.
func TestLimitUno(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)
	lim := NuevoLimitOperator(scan, 1)
	defer lim.Close()

	filas := recolectarFilas(t, lim)
	if len(filas) != 1 {
		t.Errorf("se esperaba 1 fila con LIMIT 1 y se obtuvieron %d", len(filas))
	}
}

// TestLimitPreservaEsquema verifica que Schema() delega correctamente al hijo.
func TestLimitPreservaEsquema(t *testing.T) {
	tabla := tablaDeEmpleados()
	scan := NuevoScanOperator(tabla)
	lim := NuevoLimitOperator(scan, 1)

	esquema := lim.Schema()
	if len(esquema.Columnas) != len(tabla.Esquema.Columnas) {
		t.Errorf("el esquema del LimitOperator no coincide con el del hijo")
	}
}
