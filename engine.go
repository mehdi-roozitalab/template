package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mehdi-roozitalab/core_utils"
)

// TemplateEngine a simple engine that simplify working with templates
type TemplateEngine interface {
	PathToNameTranslator() func(string) string
	// SetPathToNameTranslator set the function that will be used to translate a path to name of the template
	SetPathToNameTranslator(translator func(path string) string) TemplateEngine

	PathToFactoryTranslator() func(string) TemplateFactory
	// SetPathToFactoryTranslator set the function that will be used to translate a path to a ``TemplateFactory``
	SetPathToFactoryTranslator(translator func(path string) TemplateFactory) TemplateEngine

	// RegisterFunction register a function to be avaiable in all templates that parsed by this engine
	RegisterFunction(name string, fn interface{}) TemplateEngine
	// RegisterVariable register a variable to be avaiable in all templates that rendered by this engine
	RegisterVariable(name string, value interface{}) TemplateEngine

	// ParseTemplate parse a template using all the functions that are available in this engine
	ParseTemplate(factory TemplateFactory, text string) (Template, error)
	// ParseNamedTemplate parse a named template adding all registered functions and variables to the parser
	ParseNamedTemplate(factory TemplateFactory, name, text string) (Template, error)

	// AddTemplateExtension add a set of extensions to list of known extensions of the engine
	AddTemplateExtension(factory TemplateFactory, ext ...string) TemplateEngine
	// AddTemplateSearchPath add a set of paths to search paths for templates
	AddTemplateSearchPath(path ...string) TemplateEngine
	// LoadTemplate load a single template from specified path
	LoadTemplate(factory TemplateFactory, path string) (Template, error)
}

type simpleTemplateEngine struct {
	funcs                   map[string]interface{}
	extensions              map[string]TemplateFactory
	searchPaths             []string
	pathToNameTranslator    func(string) string
	pathToFactoryTranslator func(string) TemplateFactory
}

func NewTemplateEngine() TemplateEngine {
	engine := &simpleTemplateEngine{}
	engine.pathToNameTranslator = engine.nameForPath
	engine.pathToFactoryTranslator = engine.factoryForPath
	return engine
}

func (e *simpleTemplateEngine) PathToNameTranslator() func(string) string {
	return e.pathToNameTranslator
}
func (e *simpleTemplateEngine) SetPathToNameTranslator(translator func(path string) string) TemplateEngine {
	e.pathToNameTranslator = translator
	return e
}

func (e *simpleTemplateEngine) PathToFactoryTranslator() func(string) TemplateFactory {
	return e.pathToFactoryTranslator
}
func (e *simpleTemplateEngine) SetPathToFactoryTranslator(translator func(string) TemplateFactory) TemplateEngine {
	e.pathToFactoryTranslator = translator
	return e
}

func (e *simpleTemplateEngine) RegisterFunction(name string, fn interface{}) TemplateEngine {
	e.funcs[name] = fn
	return e
}
func (e *simpleTemplateEngine) RegisterVariable(name string, value interface{}) TemplateEngine {
	return e.RegisterFunction(name, func() interface{} { return value })
}
func (e *simpleTemplateEngine) ParseTemplate(factory TemplateFactory, text string) (Template, error) {
	ctx := templateParseContext(e.funcs)
	return factory.Parse(ctx, text)
}
func (e *simpleTemplateEngine) ParseNamedTemplate(factory TemplateFactory, name, text string) (Template, error) {
	ctx := templateParseContext(e.funcs)
	return factory.ParseWithName(ctx, name, text)
}

func (e *simpleTemplateEngine) AddTemplateExtension(factory TemplateFactory, ext ...string) TemplateEngine {
	for _, item := range ext {
		e.extensions[item] = factory
	}
	return e
}
func (e *simpleTemplateEngine) AddTemplateSearchPath(path ...string) TemplateEngine {
	for _, item := range path {
		if !core_utils.StringArrayContains(e.searchPaths, item) {
			e.searchPaths = append(e.searchPaths, item)
		}
	}
	return e
}

func (e *simpleTemplateEngine) LoadTemplate(factory TemplateFactory, path string) (Template, error) {
	if fullpath, err := e.search(path); err != nil {
		return nil, err
	} else {
		return e.loadPath(factory, fullpath)
	}
}

func (e *simpleTemplateEngine) search(path string) (string, error) {
	if filepath.IsAbs(path) {
		if !core_utils.Exists(path) {
			return "", os.ErrNotExist
		}
		return path, nil
	}

	for _, p := range e.searchPaths {
		mergedPath := filepath.Join(p, path)
		fullpath, err := core_utils.AbsolutePath(mergedPath)
		if err != nil {
			return "", err
		}

		if core_utils.Exists(fullpath) {
			return fullpath, nil
		}

		for ext := range e.extensions {
			fullpathWithExt := fullpath + ext
			if core_utils.Exists(fullpathWithExt) {
				return fullpathWithExt, nil
			}
		}
	}

	return "", os.ErrNotExist
}
func (e *simpleTemplateEngine) loadPath(factory TemplateFactory, path string) (Template, error) {
	name := e.pathToNameTranslator(path)
	if factory == nil {
		if factory = e.pathToFactoryTranslator(path); factory == nil {
			return nil, fmt.Errorf("can't find a factory that can parse specified path(%s)", path)
		}
	}
	if t := factory.Lookup(name); t != nil {
		return t, nil
	} else if content, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		return factory.ParseWithName(templateParseContext(e.funcs), name, string(content))
	}
}
func (e *simpleTemplateEngine) nameForPath(path string) string {
	name := filepath.Base(path)
	for ext := range e.extensions {
		if strings.HasSuffix(name, ext) {
			return name[:len(name)-len(ext)]
		}
	}
	return name
}
func (e *simpleTemplateEngine) factoryForPath(path string) TemplateFactory {
	ext := filepath.Ext(path)
	return e.extensions[ext]
}

type threadSafeTemplateEngine struct {
	m      sync.Mutex
	engine TemplateEngine
}

func NewThreadSafeTemplateEngine(engine TemplateEngine) TemplateEngine {
	return &threadSafeTemplateEngine{engine: engine}
}

func (e *threadSafeTemplateEngine) PathToNameTranslator() func(string) string {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.PathToNameTranslator()
}
func (e *threadSafeTemplateEngine) SetPathToNameTranslator(translator func(path string) string) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.SetPathToNameTranslator(translator)
}
func (e *threadSafeTemplateEngine) PathToFactoryTranslator() func(string) TemplateFactory {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.PathToFactoryTranslator()
}
func (e *threadSafeTemplateEngine) SetPathToFactoryTranslator(translator func(path string) TemplateFactory) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.SetPathToFactoryTranslator(translator)
}
func (e *threadSafeTemplateEngine) RegisterFunction(name string, fn interface{}) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.RegisterFunction(name, fn)
}
func (e *threadSafeTemplateEngine) RegisterVariable(name string, value interface{}) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.RegisterVariable(name, value)
}
func (e *threadSafeTemplateEngine) ParseTemplate(factory TemplateFactory, text string) (Template, error) {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.ParseTemplate(factory, text)
}
func (e *threadSafeTemplateEngine) ParseNamedTemplate(factory TemplateFactory, name, text string) (Template, error) {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.ParseNamedTemplate(factory, name, text)
}
func (e *threadSafeTemplateEngine) AddTemplateExtension(factory TemplateFactory, ext ...string) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.AddTemplateExtension(factory, ext...)
}
func (e *threadSafeTemplateEngine) AddTemplateSearchPath(path ...string) TemplateEngine {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.AddTemplateSearchPath(path...)
}
func (e *threadSafeTemplateEngine) LoadTemplate(factory TemplateFactory, path string) (Template, error) {
	e.m.Lock()
	defer e.m.Unlock()

	return e.engine.LoadTemplate(factory, path)
}

var globalTemplateEngine TemplateEngine = NewThreadSafeTemplateEngine(NewTemplateEngine().
	AddTemplateSearchPath(filepath.Join(appFolder, "templates")).
	AddTemplateSearchPath(filepath.Join(startDir, "templates")).
	AddTemplateSearchPath("./templates").
	AddTemplateExtension(TextTemplateFactory(), ".gotmpl").
	AddTemplateExtension(HtmlTemplateFactory(), ".gohtmpl").
	RegisterFunction("HOST_NAME", os.Hostname).
	RegisterFunction("SHELL", func() string {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		return shell
	}).
	RegisterVariable("START_DIR", startDir).
	RegisterVariable("APP_LOCATION", appLocation).
	RegisterVariable("APP_FOLDER", appFolder),
)

func GlobalTemplateEngine() TemplateEngine { return globalTemplateEngine }

func RegisterTemplateFunction(name string, fn interface{}) TemplateEngine {
	return GlobalTemplateEngine().RegisterFunction(name, fn)
}
func RegisterTemplateVariable(name string, val interface{}) TemplateEngine {
	return GlobalTemplateEngine().RegisterVariable(name, val)
}

func ParseTemplate(factory TemplateFactory, text string) (Template, error) {
	return GlobalTemplateEngine().ParseTemplate(factory, text)
}
func ParseTextTemplate(text string) (Template, error) {
	return ParseTemplate(TextTemplateFactory(), text)
}
func ParseHtmlTemplate(text string) (Template, error) {
	return ParseTemplate(HtmlTemplateFactory(), text)
}

func ParseNamedTemplate(factory TemplateFactory, name, text string) (Template, error) {
	return GlobalTemplateEngine().ParseNamedTemplate(factory, name, text)
}
func ParseNamedTextTemplate(name, text string) (Template, error) {
	return ParseNamedTemplate(TextTemplateFactory(), name, text)
}
func ParseHtmlNamedTemplate(name, text string) (Template, error) {
	return ParseNamedTemplate(HtmlTemplateFactory(), name, text)
}

func AddTemplateExtension(factory TemplateFactory, ext ...string) TemplateEngine {
	return GlobalTemplateEngine().AddTemplateExtension(factory, ext...)
}
func AddTemplateSearchPath(path ...string) TemplateEngine {
	return GlobalTemplateEngine().AddTemplateSearchPath(path...)
}
func LoadTemplate(factory TemplateFactory, path string) (Template, error) {
	return GlobalTemplateEngine().LoadTemplate(factory, path)
}
