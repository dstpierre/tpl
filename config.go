package tpl

// Option is used to configure some behavior and turn on features.
type Option struct {
	// TemplateRootName to override the default '/templates' root path.
	TemplateRootName string

	// EnableStaticAnalysis when enable, it will generate an AST for all
	// templates when you call Render
	EnableStaticAnalysis bool
	// StaticAnalysisFile is the filename where the AST is saved
	// so the tpl CLI can perform static analysis.
	StaticAnalysisFile string
}

var config Option

func init() {
	config = Option{
		TemplateRootName: "templates",
	}
}

// Set overrides the default option. By default the template root name is
// `templates`.
func Set(opts Option) {
	config = opts
}
