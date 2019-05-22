package goweb

const (
	// Flags used in the application.
	FlagSource = "src"
	FlagTemplatePathRepositoryInterface = "interfaceTemplate"
	FlagTemplatePathRepositoryImplementation = "implementationTemplate"

	// Constants used to parse the annotations.
	AnnotationPrefix = "goweb:"
	ModelAnnotation  = AnnotationPrefix + "Model"

	// Errors used on generators.
	ErrInvalidGoFile = "source is not a valid go file"

	// Where to find the templates to use for generated code.
	TemplatePathRepositoryInterface      = "https://raw.githubusercontent.com/masseelch/gowebapp/master/templates/repository_interface.gotpl"
	TemplatePathRepositoryImplementation = "https://raw.githubusercontent.com/masseelch/gowebapp/master/templates/repository_implementation.gotpl"
)
