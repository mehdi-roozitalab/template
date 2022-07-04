package template

import (
	"io"
)

// Template generalization of Template concept in go(text|html)
type Template interface {
	Render(data interface{}) (string, error)
	RenderTo(w io.Writer, data interface{}) error
}
