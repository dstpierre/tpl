package tpl

type Option struct {
	TemplateRootName string
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
