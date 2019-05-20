package gowebapp

import "flag"

type RepositoryFlags struct {
	ModelPackage               string
	InterfaceTemplatePath      string
	ImplementationTemplatePath string
	Src                        string
}

type Flags struct {
	Repository RepositoryFlags
}

func ParseFlags() Flags {
	var fs Flags

	flag.StringVar(&fs.Repository.ModelPackage, "repository:package", "app", "Name of package to use for generated interfaces.")
	flag.StringVar(&fs.Repository.InterfaceTemplatePath, "repository:templates:interface", "https://raw.githubusercontent.com/masseelch/gowebapp/master/templates/repository_interface.gotpl", "Path to the repository-interface template.")
	flag.StringVar(&fs.Repository.ImplementationTemplatePath, "repository:templates:implementation", "https://raw.githubusercontent.com/masseelch/gowebapp/master/templates/repository_implementation.gotpl", "Path to the repository-implementation template.")
	flag.StringVar(&fs.Repository.Src, "repository:source", "", "Path to source-file containing model-declarations.")

	flag.Parse()

	return fs
}
