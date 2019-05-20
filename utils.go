package gowebapp

import (
	"github.com/jinzhu/inflection"
	"regexp"
	"strings"
	"text/template"
)

const (
	KAnnotationPrefix = "generator:"
	KModelAnnotation  = KAnnotationPrefix + "Model"
)

var (
	matchGoFile          = regexp.MustCompile("(.)\\.go$")
	matchGeneratedGoFile = regexp.MustCompile("(.)\\.g\\.go$")
	funcMap              = template.FuncMap{
		"ToLower": strings.ToLower,
		"Plural":  inflection.Plural,
	}
)
