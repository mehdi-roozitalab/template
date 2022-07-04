package template

import (
	"html/template"
	"io"
)

type HtmlTemplate struct {
	*template.Template
}

func (t HtmlTemplate) Render(data interface{}) (string, error) { return RenderToStringImpl(t, data) }
func (t HtmlTemplate) RenderTo(w io.Writer, data interface{}) error {
	return t.Template.Execute(w, data)
}

type htmlTemplateFactory struct {
	root *template.Template
}

func (f *htmlTemplateFactory) Parse(ctx TemplateParseContext, text string) (Template, error) {
	return f.ParseWithName(ctx, "", text)
}
func (f *htmlTemplateFactory) ParseWithName(ctx TemplateParseContext, name, text string) (Template, error) {
	if t, err := f.root.New(name).Funcs(ctx.Funcs()).Parse(text); err != nil {
		return nil, err
	} else {
		return HtmlTemplate{t}, nil
	}
}
func (f *htmlTemplateFactory) Lookup(name string) Template {
	if t := f.root.Lookup(name); t != nil {
		return &HtmlTemplate{t}
	} else {
		return nil
	}
}

var _htmlTemplateFactory TemplateFactory = &htmlTemplateFactory{}

func HtmlTemplateFactory() TemplateFactory { return _htmlTemplateFactory }
