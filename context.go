package template

type TemplateParseContext interface {
	Funcs() map[string]interface{}
}

type templateParseContext map[string]interface{}

func (ctx templateParseContext) Funcs() map[string]interface{} {
	return map[string]interface{}(ctx)
}
