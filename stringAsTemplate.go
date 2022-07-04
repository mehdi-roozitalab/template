package template

import "io"

type StringAsTemplate string

func (t StringAsTemplate) Render(data interface{}) (string, error) { return string(t), nil }
func (t StringAsTemplate) RenderTo(w io.Writer, data interface{}) error {
	_, err := w.Write([]byte(t))
	return err
}
