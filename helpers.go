package template

import "strings"

// RenderToStringImpl helper function to simplify implementing Template.Render
func RenderToStringImpl(t Template, data interface{}) (string, error) {
	builder := &strings.Builder{}
	if err := t.RenderTo(builder, data); err != nil {
		return "", err
	} else {
		return builder.String(), nil
	}
}

// RegisterTemplateVariableImpl helper function to simplify implementing TemplateEngine.RegisterTemplateVariable
func RegisterTemplateVariableImpl(engine TemplateEngine, name string, val interface{}) TemplateEngine {
	return engine.RegisterFunction(name, func() interface{} { return val })
}
