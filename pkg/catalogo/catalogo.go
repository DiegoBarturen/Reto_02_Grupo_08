package catalogo

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"reto_02_grupo_08/pkg/almacenamiento"
)

type Catalogo struct {
	tablas map[string]*almacenamiento.Tabla
}

func Nuevo() *Catalogo {
	return &Catalogo{
		tablas: make(map[string]*almacenamiento.Tabla),
	}
}

func (c *Catalogo) RegistrarTabla(tabla *almacenamiento.Tabla) error {
	if c == nil {
		return errors.New("el catálogo no está inicializado")
	}

	if tabla == nil {
		return errors.New("no se puede registrar una tabla nula")
	}

	nombreTabla := strings.TrimSpace(tabla.Nombre)

	if nombreTabla == "" {
		return errors.New("el nombre de la tabla no puede estar vacío")
	}

	if c.tablas == nil {
		c.tablas = make(map[string]*almacenamiento.Tabla)
	}

	nombreNormalizado := normalizarNombre(nombreTabla)

	if _, existe := c.tablas[nombreNormalizado]; existe {
		return fmt.Errorf(
			"la tabla %q ya está registrada",
			nombreTabla,
		)
	}

	tabla.Nombre = nombreTabla
	c.tablas[nombreNormalizado] = tabla

	return nil
}

func (c *Catalogo) ObtenerTabla(nombre string) (*almacenamiento.Tabla, error) {
	if c == nil {
		return nil, errors.New("el catálogo no está inicializado")
	}

	nombreNormalizado := normalizarNombre(nombre)

	if nombreNormalizado == "" {
		return nil, errors.New("el nombre de la tabla no puede estar vacío")
	}

	tabla, existe := c.tablas[nombreNormalizado]

	if !existe {
		return nil, fmt.Errorf(
			"la tabla %q no existe en el catálogo",
			strings.TrimSpace(nombre),
		)
	}

	return tabla, nil
}

func (c *Catalogo) ExisteTabla(nombre string) bool {
	if c == nil {
		return false
	}

	_, existe := c.tablas[normalizarNombre(nombre)]
	return existe
}

func (c *Catalogo) ListarTablas() []string {
	if c == nil {
		return []string{}
	}

	nombres := make([]string, 0, len(c.tablas))

	for _, tabla := range c.tablas {
		nombres = append(nombres, tabla.Nombre)
	}

	sort.Strings(nombres)

	return nombres
}

func (c *Catalogo) ObtenerEsquema(
	nombreTabla string,
) (almacenamiento.Esquema, error) {
	tabla, err := c.ObtenerTabla(nombreTabla)
	if err != nil {
		return almacenamiento.Esquema{}, err
	}

	columnas := make(
		[]almacenamiento.Columna,
		len(tabla.Esquema.Columnas),
	)

	copy(columnas, tabla.Esquema.Columnas)

	return almacenamiento.Esquema{
		Columnas: columnas,
	}, nil
}

func normalizarNombre(nombre string) string {
	return strings.ToLower(strings.TrimSpace(nombre))
}
