// package tpl handles structuring, parsing, and rendering of HTML templates.
// It adds helpers, translations and internationalization functions.
//
// You must adopt the following structure for your templates:
// Create a directory named "templates" and create the following structure.
//
// templates/emails
//
//	templates/partials
//	templates/views/layout-name/page-name.html
//	templates/translations/en.json
//	templates/layout-name.html
//
// You create your base layouts at the root of the templates directory.
//
// Each layout must have a views/[layout_name_without_html] directory.
//
// Inside your layout files you define blocks that are filled from the views.
// For example, in your layout:
//
//	<title>{{ block "title" . }}</title>
//	<main>{{ block "content" .}}</main>
//
// And inside your view in the templates/views/[layout name]/[view name].html:
//
//	{{define "title"}}Title from the view{{end}}
//	{{define "content"}}
//	<h1>Hello from view</h1>
//	{{end}}
//
// You'll need to call the `Parse` function when your program starts and
// provide a `embed.FS` for your templates.
//
//	//go:embed templates
//	var fs embed.FS
//
//	var templ *tpl.Template
//
//	func main() {
//	  templ, err := tpl.Parse(fs, nil)
//	}
//
// When rendering a view you can optionally use the `PageData` structure or your own.
//
//	func hello(w http.ResponseWriter, r *http.Request) {
//	  data := tpl.PageData{
//	    Lang: "fr", // if needed, should match fr.json in translations dir
//	    Locale: "fr-CA", // used to format dates and currency
//	    Title: "Page title", // if you need this
//	    CurrentUser: YourUser{}, // a handy field to hold the current user
//	    Data: YourData{}, // this is what you'd normally sent to the Execute fn
//	  }
//	  if err := templ.Render(w, "app/hello.html", data); err != nil {}
//	}
//
// If you need to translate your template you may create JSON files per language
// with the following structure:
//
//	[{
//	  "key": "a unique key",
//	  "value": "translation value",
//	  "plural": "optional if plural is needed",
//	}]
//
// There's four different template function relative to translation:
//
// 1. {{ t .Lang "a unique key" }}
//
// 2. {{ tp .Lang "single or plural" 2 }}
//
// 3. {{ tf .Lang "a formatted" .Data.AnArray }}
//
// 4. {{ tpf .Lang "foramtted and pluralized" 2 .Data.AnArray }}
package tpl

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"path"
	"path/filepath"
	"strings"
)

// Template holds the file system and the parsed views.
type Template struct {
	FS     embed.FS
	Views  map[string]*template.Template
	Emails map[string]*template.Template
}

// Parse parses and load the layouts, templates, partials, and optionally the
// translation files.
//
// You should embed the templates in your program and pass the `embed.FS` to the
// function.
func Parse(fs embed.FS, funcMap map[string]any) (*Template, error) {
	if funcMap == nil {
		funcMap = make(map[string]any)
	}

	enhanceFuncMap(funcMap)

	if err := loadTranslations(fs); err != nil {
		return nil, err
	}

	partials, err := load(fs, config.TemplateRootName, "partials")
	if err != nil {
		return nil, err
	}

	layouts, err := load(fs, config.TemplateRootName)
	if err != nil {
		return nil, err
	}

	viewsDir := path.Join(config.TemplateRootName, "views")
	views := make(map[string]*template.Template)

	for _, layout := range layouts {
		layoutView := strings.TrimSuffix(layout.name, filepath.Ext(layout.name))

		pages, err := load(fs, viewsDir, layoutView)
		if err != nil {
			return nil, err
		}

		for _, view := range pages {
			viewName := fmt.Sprintf(layoutView+"/%s", view.name)

			tf := template.New(layout.name).Funcs(funcMap)

			patterns := []string{
				layout.fullPath,
				view.fullPath,
			}

			patterns = append(patterns, getPaths(partials)...)

			t, err := tf.ParseFS(
				fs,
				patterns...,
			)
			if err != nil {
				return nil, err
			}

			views[viewName] = t
		}
	}

	emails := make(map[string]*template.Template)

	emailFiles, err := load(fs, config.TemplateRootName, "emails")
	if err != nil {
		return nil, err
	}

	for _, ef := range emailFiles {
		t, err := template.New(ef.name).Funcs(funcMap).ParseFS(fs, ef.fullPath)
		if err != nil {
			return nil, err
		}

		emails[ef.name] = t
	}

	templ := &Template{FS: fs, Views: views, Emails: emails}
	return templ, nil
}

type file struct {
	name     string
	fullPath string
}

func load(fs embed.FS, dir ...string) ([]file, error) {
	var files []file

	fullDir := path.Join(dir...)

	if ok := exists(fs, fullDir); !ok {
		if strings.HasSuffix(fullDir, "_partials") {
			fmt.Println("tpl: You must have a `partials` directory created")
		} else if strings.HasSuffix(fullDir, "partials") {
			fmt.Println("tpl: obsolete name '_partials' must be changed to 'partials'.")
			dir[len(dir)-1] = "_partials"
			return load(fs, dir...)
		}

		return nil, nil
	}

	//TODO: might be an idea to un-hardcode the paths and have options
	allFiles, err := fs.ReadDir(fullDir)
	if err != nil {
		return nil, err
	}

	for _, f := range allFiles {
		if f.IsDir() {
			continue
		}

		files = append(files, file{name: f.Name(), fullPath: path.Join(fullDir, f.Name())})
	}

	return files, nil
}

func getPaths(files []file) []string {
	var p []string
	for _, f := range files {
		p = append(p, f.fullPath)
	}
	return p
}

type Notification struct {
	Title     template.HTML
	Message   template.HTML
	IsSuccess bool
	IsError   bool
	IsWarning bool
}

type PageData struct {
	Lang     string
	Locale   string
	Timezone string

	XSRFToken string

	Title       string
	CurrentUser any
	Alert       *Notification
	Data        any
	Extra       any

	Env string
}

// Render renders a template from a [layout]/[page.html].
//
// The layout should not have the .html, so if you have 2 layouts one name
// layout.html and one named app.html, a template named "dashboard.html" in the
// app layout would be named: app/dashboard.html.
func (templ *Template) Render(w io.Writer, view string, data any) error {
	v, ok := templ.Views[view]
	if !ok {
		return errors.New("can't find view: " + view)
	}

	return v.Execute(w, data)
}

// RenderEmail renders the email found in the templates/emails directory.
//
// You may create language specific templates and html and text version
// as follow: templates/emails/verify_en.html, templates/emails/verify_fr.txt, etc.
//
// Note that this execution does not use the PageData struct, but the data
// passed directly.
func (templ *Template) RenderEmail(w io.Writer, email string, data any) error {
	e, ok := templ.Emails[email]
	if !ok {
		return errors.New("can't find email: " + email)
	}

	return e.Execute(w, data)
}

// exists returns whether the given file or directory exists
func exists(fs embed.FS, path string) bool {
	f, err := fs.Open(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// GetDataContent returns the content of file in the data directory
func (templ *Template) GetDataContent(filename string) ([]byte, error) {
	return templ.FS.ReadFile(path.Join(config.TemplateRootName, "data", filename))
}
