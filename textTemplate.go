package template

import (
	"io"
	"text/template"
)

type TextTemplate struct {
	*template.Template
}

func (t TextTemplate) Render(data interface{}) (string, error) { return RenderToStringImpl(t, data) }
func (t TextTemplate) RenderTo(w io.Writer, data interface{}) error {
	return t.Template.Execute(w, data)
}

type textTemplateFactory struct {
	root *template.Template
}

func (f *textTemplateFactory) Parse(ctx TemplateParseContext, text string) (Template, error) {
	return f.ParseWithName(ctx, "", text)
}
func (f *textTemplateFactory) ParseWithName(ctx TemplateParseContext, name, text string) (Template, error) {
	if t, err := f.root.New(name).Funcs(ctx.Funcs()).Parse(text); err != nil {
		return nil, err
	} else {
		return TextTemplate{t}, nil
	}
}
func (f *textTemplateFactory) Lookup(name string) Template {
	if t := f.root.Lookup(name); t != nil {
		return TextTemplate{t}
	} else {
		return nil
	}
}

var _textTemplateFactory TemplateFactory = &textTemplateFactory{}

func TextTemplateFactory() TemplateFactory { return _textTemplateFactory }
