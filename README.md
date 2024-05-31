# tpl

![build badge](https://github.com/dstpierre/tpl/actions/workflows/test.yml/badge.svg)
[![GoReportCard](https://goreportcard.com/badge/github.com/dstpierre/tpl)](https://goreportcard.com/report/github.com/dstpierre/tpl)
[![Go Reference](https://pkg.go.dev/badge/github.com/dstpierre/tpl.svg)](https://pkg.go.dev/github.com/dstpierre/tpl)


> I was tired of having the same issues phase when starting a new web project with Go's `html/template`.

`tpl` is an opinionated lightweight helper library that makes structuring, parsing, and rendering templates in Go more tolerable. It also adds i18n helpers to the `funcmap` of templates.

**Table of content**:

* [Installation](#installation)
* [Usage](#usage)
  * [Template file structure](#template-file-structure)
  * [Parsing and rendering](#parsing-and-rendering)
  * [PageData structure](#pageData-structure)
* [Example templates](#example-templates)
  * [Quick template example](#quick-template-example)
* [i18n](#i18n)
* [Passing a funcmap](#passing-a-funcmap)

## Installation

```sh
$ go get github.com/dstpierre/tpl
```

## Usage

### Template file structure

To use this library, you'll need to adopt the following files and directory structure for your templates:

Create a `templates` directory with the following structure:

```
templates/
├── _partials
│   └── a-reusable-piece-1.html
│   └── a-reusable-piece-2.html
├── app.html
├── layout.html
├── translations
│   ├── en.json
│   └── fr.json
└── views
    ├── app
    │   ├── dashboard.html
    │   └── page-signed-in-user.html
    └── layout
        └── user-login.html
```

Now `app.html` and `layout.html` are **example names**.

**Layouts** are HTML files at the root of your `templates/` directory. They contain blocks that your views will fill.

**views** directory contains one directory per layout file name without the .html extension. If you have three layout templates, `home.html`, `blog.html`, and `another.html`, you'll have three sub-directories in the Views directory, each containing the views for this layout.

**_partials** is a directory where you put all re-usable pieces of template you need to embed into your HTML pages. For instance, you embed a `blog-view.html` in 'views/blog/list.html', `views/blog/category.html`, and `views/blog/tag.html` pages.

**translations** directory is where you put message translations via one file named after the language. It's optional.

### Parsing and rendering

You'll need to parse your templates at the start of your program. The library returns a `tpl.Template` structure that you use to render your pages.

For example:

```go
package main

import (
  "embed"
  "net/http"
  "github.com/dstpierre/tpl"
)

//go:embed templates/*
var fs embed.FS

func main() {
  // assuming your templates are in templates/ and have proper structure
  templ, err := tpl.Parse(fs, nil)
  // ...
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request)) {
    data := "this is your app's data you normally pass to Execute"
    pdata := tpl.PageData{Data: data}
    if err := templ.Render(w, "app/dashboard.html", pdata); err != nil {}
  }
}
)
```

As you can see, you need to wrap your data inside a `tpl.PageData` structure. This enables the library to perform lingual translations and internationalize dates and currencies.

### PageData structure

Here's the fields of the `tpl.PageData`:

```go
type PageData struct {
  Lang   string
  Locale string
  Title       string
  CurrentUser any
  Data        any
}
```

`Lang` and `Locale` are useful if you want to use the i18n feature.

`CurrentUser` is handy if you want to let your templates know about the current user.

`Title` is also helpful to set the page title, you can have this in your layout templates:

```html
<title>{{.Title}}</title>
or
{{if .Title}}
  {{.Title}}
{{else}}
  Default title when empty
{{end}}
```

## Example templates

The tests use somewhat real-ish directories and file structures from which you may get started.

Look at the `testdata` directory. In your program, you might want to name the root directory `templates` but it's configurable.

### Quick template example

**templates/layout.html**:

```html
{{template "nav.html"}}

<main>{{block "content" .}}{{end}}</main>
```

**templates/views/layout/home.html**:

```html
{{define "content"}}
<h1>From the home.html view</h1>
{{end}}
```

**templates/_partials/nav.html**:

```html
<nav>
  Navigation would goes here
</nav>
```

## i18n

If your web application needs multilingual support, you can create language message files and save them in the Translations directory.

**templates/translations/en.json**:

```json
[{
  "key": "unique key",
  "value": "The value",
  "plural": "Optional value for plural",
}]
```

For the translation to work you need to set the `Lang` field of the `tpl.PageData` when rendering your template:

```go
func home(w http.ResponseWriter, r *http.Request) {
  pdata := tpl.PageData{Lang: "fr", Data: 1234}
  if err := templ.Render(w, "layout/home.html", pdata); err != nil {}
}
```

Inside your templates:

```html
<p>{{ t .Lang "unique key" }}</p>

Or for plural

<p>{{ tp .Lang "unique key" .Data }}</p>

.Data is 1234 in example above, so the plural value would be displayed.
```

There's helper function to display dates and currencies in the proper format based on `Locale`.

```go
func home(w http.ResponseWriter, r *http.Request) {
  pdata := tpl.PageData{Lang: "fr", Locale: "fr-CA", Data: 59.99}
  if err := templ.Render(w, "layout/home.html", pdata); err != nil {}
}
```

And inside the **templates/views/layout/home.html** file:

```html
<p>The price is {{ currency .Locale .Data }}</p>
```

Display: The price is 59.99 $

If `Locale` is `en-US`: The price is $55.99.

There's also a `{{ shortdate .Locale .Data.CreatedAt }}` helper function which formats a `time.Time` properly based on `Locale`.

*NOTE: At this time there's only a limited amount of locale supported. If your locale isn't supported, please consider contributing the changes.* 

## Passing a funcmap

You may have helper functions you'd like to pass to the templates. Here's how:

```go
package main

import (
  "embed"
  "github.com/dstpierre/tpl"
)

//go:embed templates/*
var fs embed.FS

var templ *tpl.Template

func main() {
  fmap := make(map[string]any)
  fmap["myfunc"] = func() string { return "hello" }
  r, err := tpl.Parse(fs, fmap)
  //...
  templ = t
}
```
