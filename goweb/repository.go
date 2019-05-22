package goweb

import (
	"bytes"
	"errors"
	"github.com/urfave/cli"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
)

type RepositoryGenerator struct {
	InterfaceTemplate      *template.Template
	ImplementationTemplate *template.Template

	// The path to the current source-file.
	Source string

	// Since the RepositoryGenerator satisfies the ast.Visitor-interface
	// we cache some information about the ast while walking through it.
	GenDecl *ast.GenDecl
	Package string
}

func (g RepositoryGenerator) generateImplementation(t *ast.TypeSpec) error {
	d := struct {
		Package            string
		Timestamp          time.Time
		Type               string
		InsertFields       []string
		InsertValues       []interface{}
		SelectDestinations []interface{}
		SelectFields       []string
		UpdateFields       []string
		UpdateValues       []interface{}
	}{Package: g.Package, Timestamp: time.Now(), Type: t.Name.Name}

	for _, field := range t.Type.(*ast.StructType).Fields.List {
		if c := field.Comment.Text(); strings.HasPrefix(c, "gen:") {
			if strings.Contains(c, "select") {
				d.SelectFields = append(d.SelectFields, toSnakeCase(field.Names[0].Name))
				d.SelectDestinations = append(d.SelectDestinations, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				d.InsertFields = append(d.InsertFields, toSnakeCase(field.Names[0].Name))
				d.InsertValues = append(d.InsertValues, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				d.UpdateFields = append(d.UpdateFields, toSnakeCase(field.Names[0].Name))
				d.UpdateValues = append(d.UpdateValues, field.Names[0].Name)
			}
		}
	}

	var buf bytes.Buffer
	err := g.ImplementationTemplate.Execute(&buf, d)
	if err != nil {
		return err
	}

	// Format generated code.
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// Write code to file.
	err = os.MkdirAll(filepath.Join(filepath.Dir(g.Source), "sql"), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(filepath.Dir(g.Source), "sql", matchGoFile.ReplaceAllString(filepath.Base(g.Source), "${1}.g.go")))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(formatted)

	return err
}

func (g RepositoryGenerator) generateInterface(t *ast.TypeSpec) error {
	var buf bytes.Buffer
	err := g.InterfaceTemplate.Execute(&buf, struct {
		Package   string
		Timestamp time.Time
		Type      string
	}{g.Package, time.Now(), t.Name.Name})
	if err != nil {
		return err
	}

	// Format generated code.
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// Write interfaces to file.
	f, err := os.Create(matchGoFile.ReplaceAllString(g.Source, "${1}.g.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(formatted)

	return err
}

func GenerateRepository(c *cli.Context) error {
	// If there is no source given abort.
	if c.String(FlagSource) == "" {
		return errors.New("no source file given")
	}

	// Get the absolute path to the given model-file.
	abs, err := filepath.Abs(c.String(FlagSource))
	if err != nil {
		return err
	}

	// If the source is a directory abort.
	src, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if src.IsDir() {
		return errors.New(ErrInvalidGoFile)
	}

	intTpl, err := parseTemplate(c.String(FlagTemplatePathRepositoryInterface))
	if err != nil {
		return err
	}

	implTpl, err := parseTemplate(c.String(FlagTemplatePathRepositoryInterface))
	if err != nil {
		return err
	}

	// Create the generator.
	g := RepositoryGenerator{
		InterfaceTemplate:      intTpl,
		ImplementationTemplate: implTpl,
		Source:                 abs,
	}

	// Parse the file.
	parsed, err := parser.ParseFile(token.NewFileSet(), g.Source, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Walk(g, parsed)

	return nil
}

// todo - parse the annotation to a struct
func (g RepositoryGenerator) getCommentGroup(spec *ast.TypeSpec) *ast.CommentGroup {
	if spec.Doc == nil {
		return g.GenDecl.Doc
	}

	return spec.Doc
}

func (g RepositoryGenerator) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch t := n.(type) {
	case *ast.GenDecl:
		g.GenDecl = t
	case *ast.TypeSpec:
		// If the type declaration is marked as a model generate repository for it.
		cg := g.getCommentGroup(t)
		if strings.HasPrefix(cg.Text(), ModelAnnotation) {
			if _, ok := t.Type.(*ast.StructType); ok {

				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					if err := g.generateInterface(t); err != nil {
						panic(err)
					}
				}()

				go func() {
					defer wg.Done()
					if err := g.generateImplementation(t); err != nil {
						panic(err)
					}
				}()

				wg.Wait()
			}
		}
	}

	return g
}
