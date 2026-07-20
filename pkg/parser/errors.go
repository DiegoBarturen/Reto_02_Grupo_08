package parser

import "fmt"

type SyntaxError struct {
	Message  string
	Position Position
}

func (e *SyntaxError) Error() string {
	if e == nil {
		return ""
	}

	return fmt.Sprintf(
		"error de sintaxis en %s: %s",
		e.Position.String(),
		e.Message,
	)
}
