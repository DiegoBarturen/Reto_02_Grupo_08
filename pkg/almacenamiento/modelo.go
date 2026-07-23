package almacenamiento

type TipoDato string

const (
	TipoEntero   TipoDato = "ENTERO"
	TipoDecimal  TipoDato = "DECIMAL"
	TipoTexto    TipoDato = "TEXTO"
	TipoBooleano TipoDato = "BOOLEANO"
)

type Valor struct {
	Tipo   TipoDato
	Dato   any
	EsNulo bool
}

type Columna struct {
	Nombre string
	Tipo   TipoDato
}

type Esquema struct {
	Columnas []Columna
}

type Fila struct {
	Datos []Valor
}

type Tabla struct {
	Nombre  string
	Esquema Esquema
	Filas   []Fila
}
