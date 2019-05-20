package gowebapp

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"net/http"
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
	ModelPackage           string

	// The path to the current source-file.
	SrcFilepath string

	// Since the RepositoryGenerator satisfies the ast.Visitor-interface
	// we cache some information about the ast while walking through it.
	GenDecl *ast.GenDecl
}

func (g RepositoryGenerator) generateImplementation(t *ast.TypeSpec) {
	d := struct {
		ModelPackage       string
		Timestamp          time.Time
		Type               string
		InsertFields       []string
		InsertValues       []interface{}
		SelectDestinations []interface{}
		SelectFields       []string
		UpdateFields       []string
		UpdateValues       []interface{}
	}{ModelPackage: g.ModelPackage, Timestamp: time.Now(), Type: t.Name.Name}

	for _, field := range t.Type.(*ast.StructType).Fields.List {
		if c := field.Comment.Text(); strings.HasPrefix(c, "gen:") {
			if strings.Contains(c, "select") {
				d.SelectFields = append(d.SelectFields, ToSnakeCase(field.Names[0].Name))
				d.SelectDestinations = append(d.SelectDestinations, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				d.InsertFields = append(d.InsertFields, ToSnakeCase(field.Names[0].Name))
				d.InsertValues = append(d.InsertValues, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				d.UpdateFields = append(d.UpdateFields, ToSnakeCase(field.Names[0].Name))
				d.UpdateValues = append(d.UpdateValues, field.Names[0].Name)
			}
		}
	}

	var buf bytes.Buffer
	err := g.ImplementationTemplate.Execute(&buf, d)
	panicOnError(err)

	// Format generated code.
	formatted, err := format.Source(buf.Bytes())
	panicOnError(err)

	// Write code to file.
	err = os.MkdirAll(filepath.Join(filepath.Dir(g.SrcFilepath), "sql"), os.ModePerm)
	panicOnError(err)
	f, err := os.Create(filepath.Join(filepath.Dir(g.SrcFilepath), "sql", matchGoFile.ReplaceAllString(filepath.Base(g.SrcFilepath), "${1}.g.go")))
	panicOnError(err)
	defer f.Close()

	_, err = f.Write(formatted)
	panicOnError(err)
}

func (g RepositoryGenerator) generateInterface(t *ast.TypeSpec) {
	var buf bytes.Buffer
	err := g.InterfaceTemplate.Execute(&buf, struct {
		Timestamp time.Time
		Type      string
		Package   string
	}{time.Now(), t.Name.Name, g.ModelPackage})
	panicOnError(err)

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

func GenerateRepositories(fs RepositoryFlags) {
	abs, err := filepath.Abs(fs.Src)
	panicOnError(err)

	g := RepositoryGenerator{
		ModelPackage:           fs.ModelPackage,
		InterfaceTemplate:      parseTemplate(fs.InterfaceTemplatePath),
		ImplementationTemplate: parseTemplate(fs.ImplementationTemplatePath),
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

	var wg sync.WaitGroup
	wg.Add(len(files))

	for _, file := range files {
		go func(file os.FileInfo) {
			defer wg.Done()

			g.SrcFilepath = abs
			if src.IsDir() {
				g.SrcFilepath = filepath.Join(g.SrcFilepath, file.Name())
			}

			// Parse the file.
			parsed, err := parser.ParseFile(token.NewFileSet(), g.SrcFilepath, nil, parser.ParseComments)
			panicOnError(err)

			ast.Walk(g, parsed)
		}(file)
	}

	wg.Wait()
}

// todo - Write a comment parser to parse generator annotations.
func (g RepositoryGenerator) getCommentGroup(spec *ast.TypeSpec) *ast.CommentGroup {
	if spec.Doc == nil {
		return g.GenDecl.Doc
	}

	return spec.Doc
}

func parseTemplate(path string) *template.Template {
	var tpl []byte
	if strings.HasPrefix(path, "http") {
		var buf bytes.Buffer
		r, err := http.Get(path)
		panicOnError(err)
		_, err = io.Copy(&buf, r.Body)
		panicOnError(err)
		defer r.Body.Close()
		tpl = buf.Bytes()
	} else {
		var err error
		tpl, err = ioutil.ReadFile(path)
		panicOnError(err)
	}
	return template.Must(template.New("").Funcs(funcMap).Parse(KGeneratedFileWarningComment + string(tpl)))
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
			if _, ok := t.Type.(*ast.StructType); ok {

				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					g.generateInterface(t)
				}()

				go func() {
					defer wg.Done()
					g.generateImplementation(t)
				}()

				wg.Wait()
			}
		}
	}

	return g
}
