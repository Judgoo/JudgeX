package generator

import (
	"JudgeX/languages"
	"JudgeX/utils"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
)

// Generator is responsible for generating validation files for the given in a go source file.
type Generator struct {
	t               *template.Template
	knownTemplates  map[string]*template.Template
	fileSet         *token.FileSet
	lowercaseLookup bool
	caseInsensitive bool
	marshal         bool
	names           bool
	prefix          string
}

// Enum holds data for a discovered enum in the parsed source
type Enum struct {
	Name   string
	Prefix string
	Type   string
	Values []EnumValue
}

// EnumValue holds the individual data for each enum value within the found enum.
type EnumValue struct {
	RawName string
	Name    string
	Value   int
	Comment string
	Profile languages.LanguageProfile
}

// NewGenerator is a constructor method for creating a new Generator with default
// templates loaded.
func NewGenerator() *Generator {
	g := &Generator{
		knownTemplates:  make(map[string]*template.Template),
		t:               template.New("generator"),
		fileSet:         token.NewFileSet(),
		names:           true,
		marshal:         true,
		lowercaseLookup: true,
		caseInsensitive: true,
	}

	funcs := sprig.TxtFuncMap()

	funcs["stringifyRawName"] = StringifyRawName
	funcs["mapify"] = Mapify
	funcs["unmapify"] = Unmapify
	funcs["namify"] = Namify

	_, filename, _, _ := runtime.Caller(0)
	confPath := path.Join(filename, "../languages.tmpl")
	g.t.Funcs(funcs)
	data, err := utils.ReadDataFromFile(confPath)
	if err != nil {
		os.Exit(1)
	}
	g.t = template.Must(g.t.Parse(string(data)))

	g.updateTemplates()
	return g
}

// WithPrefix is used to add a custom prefix to the enum constants
func (g *Generator) WithPrefix(prefix string) *Generator {
	g.prefix = prefix
	return g
}

func (g *Generator) GenerateFromProfile(inputFile string, keys *[]string, profileMap *languages.LanguageProfileMap) ([]byte, error) {
	f, err := g.parseFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("generate: error parsing input file '%s': %s", inputFile, err)
	}
	pkg := f.Name.Name

	vBuff := bytes.NewBuffer([]byte{})
	err = g.t.ExecuteTemplate(vBuff, "header", map[string]interface{}{"package": pkg})
	if err != nil {
		return nil, errors.WithMessage(err, "Failed writing header")
	}

	// Parse the enum doc statement
	enum, pErr := g.parseEnum(*keys, *profileMap)
	if pErr != nil {
		return []byte{}, pErr
	}

	data := map[string]interface{}{
		"enum":      enum,
		"name":      TargetName,
		"lowercase": g.lowercaseLookup,
		"nocase":    g.caseInsensitive,
		"marshal":   g.marshal,
		"names":     g.names,
	}

	err = g.t.ExecuteTemplate(vBuff, "enum", data)
	if err != nil {
		return vBuff.Bytes(), errors.WithMessage(err, fmt.Sprintf("Failed writing enum data for enum: %q", TargetName))
	}

	formatted, err := imports.Process(pkg, vBuff.Bytes(), nil)
	if err != nil {
		err = fmt.Errorf("generate: error formatting code %s\n\n%s", err, vBuff.String())
	}
	return formatted, err
}

// updateTemplates will update the lookup map for validation checks that are
// allowed within the template engine.
func (g *Generator) updateTemplates() {
	for _, template := range g.t.Templates() {
		g.knownTemplates[template.Name()] = template
	}
}

// parseFile simply calls the go/parser ParseFile function with an empty token.FileSet
func (g *Generator) parseFile(fileName string) (*ast.File, error) {
	// Parse the file given in arguments
	return parser.ParseFile(g.fileSet, fileName, nil, parser.ParseComments)
}

// parseEnum looks for the ENUM(x,y,z) formatted documentation from the type definition
func (g *Generator) parseEnum(keys []string, profileMap languages.LanguageProfileMap) (*Enum, error) {
	enum := &Enum{}
	enum.Name = TargetName
	enum.Type = TargetType
	if g.prefix != "" {
		enum.Prefix = g.prefix + enum.Prefix
	}

	values := keys
	data := 0
	for _, value := range values {
		profile := profileMap[value]
		// Make sure to leave out any empty parts
		if value != "" {
			value := strings.TrimSpace(value)
			rawName := value
			name := strings.Title(rawName)
			ev := EnumValue{Name: name, RawName: rawName, Value: data, Profile: *profile}
			enum.Values = append(enum.Values, ev)
			data++
		}
	}
	return enum, nil
}
