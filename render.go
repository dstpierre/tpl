// package tpl handles structuring, parsing, and rendering of HTML templates.
// It adds translations and internationalization functions.
//
// You must adopt the following structure for your templates:
//
//	templates/_partials
//	templates/views/layout-name/page-name.html
//	templates/translations/en.json
//	templates/layout-name.html
//
// Inside your layout files you define blocks that are filled from the view.
// For example:
//
//	<title>{{ block "title" . }}</title>
//	<main>{{ block "content" .}}</main>
//
// And inside your view inside the /templates/views/[layout name]/[view name].html:
//
//	{{define "title"}}Title from the view{{end}}
//	{{define "content"}}
//	<h1>Hello from view</h1>
//	{{end}}
//
// You'll need to call the `Parse` function when your program starts and
// provide a `embed.FS` for your templates.
//
//	//go:embed templates/*
//	var fs embed.FS
//
//	var templ *tpl.Template
//
//	func main() {
//	  templ, err := tpl.Parse(fs)
//	}
//
// And you need to use the `PageData` structure to render a template.
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
// There's four different template function helpers:
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
	FS    embed.FS
	Views map[string]*template.Template
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

	partials, err := load(fs, config.TemplateRootName, "_partials")
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

			//fmt.Println("DEBUG: ", layout.fullPath, view.fullPath)

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

	templ := &Template{FS: fs, Views: views}
	return templ, nil
}

type file struct {
	name     string
	fullPath string
}

func load(fs embed.FS, dir ...string) ([]file, error) {
	var files []file

	fullDir := path.Join(dir...)

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

type PageData struct {
	Lang   string
	Locale string

	Title       string
	CurrentUser any
	Data        any
}

// Render renders a template from a [layout]/[page.html].
//
// The layout should not have the .html, so if you have 2 layouts one name
// layout.html and one named app.html, a template named "dashboard.html" in the
// app layout would be named: app/dashboard.html.
func (templ *Template) Render(w io.Writer, view string, data PageData) error {
	v, ok := templ.Views[view]
	if !ok {
		return errors.New("can't find view: " + view)
	}

	return v.Execute(w, data)
}
