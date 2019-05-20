package gowebapp

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type RepositoryGenerator struct {
	InterfaceTemplate      *template.Template
	ImplementationTemplate *template.Template
	ModelPackage           string

	// The path to the current source-file.
	SrcFilepath string

	// Since the RepositoryGenerator satisfies the ast.Visitor-interface
	// we cache some information about the ast while walking through it.
	GenDecl *ast.GenDecl
}

func GenerateRepositories(fs RepositoryFlags) {
	abs, err := filepath.Abs(fs.Src)
	panicOnError(err)

	// Read in the templates.
	intTpl, err := ioutil.ReadFile(fs.InterfaceTemplatePath)
	panicOnError(err)
	implTpl, err := ioutil.ReadFile(fs.ImplementationTemplatePath)
	panicOnError(err)

	g := RepositoryGenerator{
		InterfaceTemplate:      template.Must(template.New("").Funcs(funcMap).Parse(string(intTpl))),
		ImplementationTemplate: template.Must(template.New("").Funcs(funcMap).Parse(string(implTpl))),
		ModelPackage:           fs.ModelPackage,
	}

	src, err := os.Stat(abs)
	panicOnError(err)

	// If the source is a directory generate repositories for all go-files containing annotated structs.
	var files []os.FileInfo
	if src.IsDir() {
		tmp, err := ioutil.ReadDir(abs)
		panicOnError(err)
		for _, f := range tmp {
			n := []byte(f.Name())
			if !f.IsDir() && matchGoFile.Match(n) && !matchGeneratedGoFile.Match(n) {
				files = append(files, f)
			}
		}
	} else {
		files = append(files, src)
	}

	for _, file := range files {
		g.SrcFilepath = abs
		if src.IsDir() {
			g.SrcFilepath = filepath.Join(g.SrcFilepath, file.Name())
		}

		// Parse the file.
		parsed, err := parser.ParseFile(token.NewFileSet(), g.SrcFilepath, nil, parser.ParseComments)
		panicOnError(err)

		ast.Walk(g, parsed)
	}
}

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
		if strings.HasPrefix(cg.Text(), KModelAnnotation) {
			if st, ok := t.Type.(*ast.StructType); ok {
				// Generate the interface.
				g.generateInterface(t)

				fmt.Printf("%v", st)
				//filename := v.InputPath
				//if v.PathInfo.IsDir() {
				//	filename = filepath.Join(filename, v.FileInfo.Name())
				//}
				//
				//v.repositories = append(v.repositories, t.Name.Name)
				//
				//// Generate the interface.
				//v.generateInterfaces(t, filename)
				//
				//// Generate the implementations.
				//v.generateImplementations(t, st, filename)
			}
		}
	}

	return g
}

func (g RepositoryGenerator) generateInterface(t *ast.TypeSpec) {
	var buf bytes.Buffer
	err := g.InterfaceTemplate.Execute(&buf, struct {
		Timestamp time.Time
		Type      string
		Package   string
	}{time.Now(), t.Name.Name, g.ModelPackage})
	panicOnError(err)

	fmt.Printf("%v", buf.String())

	// Format generated code.
	formatted, err := format.Source(buf.Bytes())
	panicOnError(err)

	// Write interfaces to file.
	f, err := os.Create(matchGoFile.ReplaceAllString(g.SrcFilepath, "${1}.g.go"))
	panicOnError(err)
	defer f.Close()

	_, err = f.Write(formatted)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
