package catalogo

import (
	"reflect"
	"testing"

	"reto_02_grupo_08/pkg/almacenamiento"
)

func crearTablaPrueba(nombre string) *almacenamiento.Tabla {
	return &almacenamiento.Tabla{
		Nombre: nombre,
		Esquema: almacenamiento.Esquema{
			Columnas: []almacenamiento.Columna{
				{
					Nombre: "id",
					Tipo:   almacenamiento.TipoTexto,
				},
				{
					Nombre: "nombre",
					Tipo:   almacenamiento.TipoTexto,
				},
			},
		},
		Filas: []almacenamiento.Fila{
			{
				Datos: []almacenamiento.Valor{
					{
						Tipo: almacenamiento.TipoTexto,
						Dato: "1",
					},
					{
						Tipo: almacenamiento.TipoTexto,
						Dato: "Ana",
					},
				},
			},
		},
	}
}

func TestRegistrarYObtenerTabla(t *testing.T) {
	catalogo := Nuevo()
	tablaEsperada := crearTablaPrueba("estudiantes")

	err := catalogo.RegistrarTabla(tablaEsperada)
	if err != nil {
		t.Fatalf(
			"RegistrarTabla devolvió un error inesperado: %v",
			err,
		)
	}

	tablaObtenida, err := catalogo.ObtenerTabla("estudiantes")
	if err != nil {
		t.Fatalf(
			"ObtenerTabla devolvió un error inesperado: %v",
			err,
		)
	}

	if tablaObtenida != tablaEsperada {
		t.Error("no se devolvió la misma tabla registrada")
	}

	if !catalogo.ExisteTabla("estudiantes") {
		t.Error("ExisteTabla debería devolver verdadero")
	}
}

func TestNombreTablaNoDistingueMayusculas(t *testing.T) {
	catalogo := Nuevo()

	err := catalogo.RegistrarTabla(
		crearTablaPrueba("Estudiantes"),
	)
	if err != nil {
		t.Fatalf("no se pudo registrar la tabla: %v", err)
	}

	nombres := []string{
		"estudiantes",
		"ESTUDIANTES",
		"Estudiantes",
		" estudiantes ",
	}

	for _, nombre := range nombres {
		t.Run(nombre, func(t *testing.T) {
			if !catalogo.ExisteTabla(nombre) {
				t.Errorf(
					"no se encontró la tabla con el nombre %q",
					nombre,
				)
			}
		})
	}
}

func TestRechazarTablaDuplicada(t *testing.T) {
	catalogo := Nuevo()

	err := catalogo.RegistrarTabla(
		crearTablaPrueba("Estudiantes"),
	)
	if err != nil {
		t.Fatalf("no se pudo registrar la primera tabla: %v", err)
	}

	err = catalogo.RegistrarTabla(
		crearTablaPrueba("estudiantes"),
	)
	if err == nil {
		t.Fatal("se esperaba un error por tabla duplicada")
	}
}

func TestObtenerEsquema(t *testing.T) {
	catalogo := Nuevo()
	tabla := crearTablaPrueba("estudiantes")

	err := catalogo.RegistrarTabla(tabla)
	if err != nil {
		t.Fatalf("no se pudo registrar la tabla: %v", err)
	}

	esquema, err := catalogo.ObtenerEsquema("estudiantes")
	if err != nil {
		t.Fatalf(
			"ObtenerEsquema devolvió un error inesperado: %v",
			err,
		)
	}

	if !reflect.DeepEqual(esquema, tabla.Esquema) {
		t.Errorf(
			"se esperaba %#v y se obtuvo %#v",
			tabla.Esquema,
			esquema,
		)
	}
}

func TestListarTablas(t *testing.T) {
	catalogo := Nuevo()

	nombres := []string{
		"productos",
		"estudiantes",
		"cursos",
	}

	for _, nombre := range nombres {
		err := catalogo.RegistrarTabla(crearTablaPrueba(nombre))
		if err != nil {
			t.Fatalf(
				"no se pudo registrar la tabla %q: %v",
				nombre,
				err,
			)
		}
	}

	esperado := []string{
		"cursos",
		"estudiantes",
		"productos",
	}

	obtenido := catalogo.ListarTablas()

	if !reflect.DeepEqual(obtenido, esperado) {
		t.Errorf(
			"se esperaba %v y se obtuvo %v",
			esperado,
			obtenido,
		)
	}
}

func TestErroresCatalogo(t *testing.T) {
	catalogo := Nuevo()

	pruebas := []struct {
		nombre string
		tabla  *almacenamiento.Tabla
	}{
		{
			nombre: "tabla nula",
			tabla:  nil,
		},
		{
			nombre: "nombre vacío",
			tabla:  crearTablaPrueba(""),
		},
		{
			nombre: "nombre con espacios",
			tabla:  crearTablaPrueba("   "),
		},
	}

	for _, prueba := range pruebas {
		t.Run(prueba.nombre, func(t *testing.T) {
			err := catalogo.RegistrarTabla(prueba.tabla)

			if err == nil {
				t.Fatal(
					"se esperaba un error al registrar la tabla",
				)
			}
		})
	}

	_, err := catalogo.ObtenerTabla("tabla_inexistente")
	if err == nil {
		t.Fatal(
			"se esperaba un error al consultar una tabla inexistente",
		)
	}
}
