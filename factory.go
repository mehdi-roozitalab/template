package template

type TemplateFactory interface {
	Lookup(name string) Template
	Parse(ctx TemplateParseContext, text string) (Template, error)
	ParseWithName(ctx TemplateParseContext, name, text string) (Template, error)
}
