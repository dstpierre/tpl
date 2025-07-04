# tpl

![build badge](https://github.com/dstpierre/tpl/actions/workflows/test.yml/badge.svg)
[![GoReportCard](https://goreportcard.com/badge/github.com/dstpierre/tpl)](https://goreportcard.com/report/github.com/dstpierre/tpl)
[![Go Reference](https://pkg.go.dev/badge/github.com/dstpierre/tpl.svg)](https://pkg.go.dev/github.com/dstpierre/tpl)


> I was tired of having the same issues phase when starting a new web project with Go's `html/template`.

`tpl` is an opinionated lightweight helper library that makes structuring, parsing, and rendering templates in Go more tolerable. It also adds small helpers, translations, and i18n functions to the `funcmap` of templates.

**Table of content**:

* [Installation](#installation)
* [Usage](#usage)
  * [Template file structure](#template-file-structure)
  * [Parsing and rendering](#parsing-and-rendering)
  * [PageData structure](#pageData-structure)
* [Example templates](#example-templates)
  * [Quick template example](#quick-template-example)
  * [Rendering emails](#rendering-emails)
* [i18n](#i18n)
* [Passing a funcmap](#passing-a-funcmap)
* [Built-in functions]()#built-in-functions

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
├── emails
│   └── verify-email.html
│   └── verify-email.txt
├── partials
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

Now `app.html` and `layout.html` are **example names**, you name your layout the way you want.

**Layouts** are HTML files at the root of your `templates/` directory. They contain blocks that your views will fill. You may name them as you want but they must have a sub-directory in the `views` directory with their name without the `.html`.

**views** directory contains one directory per layout file name without the .html extension. If you have three layout templates, `public.html`, `app.html`, and `xyz.html`, you'll have three sub-directories in the Views directory, each containing the views for this layout. So `views/public`, `views/app`, and `views/xyz`.

**partials** is a directory where you put all re-usable pieces of template you need to embed into your HTML pages. For instance, you embed a `item-list.html` in 'views/xyz/list.html', `views/app/mylist.html`, and `views/xyz/admin-list.html` pages. All three can use the partial.

__Note__: This directory was named `_partials` before, please rename it to `partials` to remove the obsolete warning.

**emails** directory is used for your emails, those HTML templates does not have a base layout and are used as-is in terms of rendering.

**translations** directory is where you put message translations via one file named after the language. It's optional, if you don't need translations you don't need to create this directory.

### Parsing and rendering

You'll need to parse your templates at the start of your program. The library returns a `tpl.Template` structure that you use to render your pages and emails.

For example:

```go
package main

import (
  "embed"
  "net/http"
  "github.com/dstpierre/tpl"
)

//go:embed templates
var fs embed.FS

func main() {
  // assuming your templates are in templates/ and have proper structure
  templ, err := tpl.Parse(fs, nil)
  // ...
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request)) {
    data := "this is your app's data you normally pass to Execute"
    if err := templ.Render(w, "app/dashboard.html", data); err != nil {}
  }
}
)
```

__Note__: Previously it was required to wrap your template data into `tpl.PageData` structure. It's not required anymore, although `tpl` still exposes `PageData` you can use or embed into your own structure.

For new project I tend to use `tpl.PageData` like this:

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request)) {
  data := "this is your app's data you normally pass to Execute"
  pdata := tpl.PageData(Data: data, Lang: "fr")
  if err := templ.Render(w, "app/dashboard.html", pdata); err != nil {}
}
```

And for existing project, I tend to embed `tpl.PageData` into my existing structure, so my HTML templates does not change much.

### PageData structure

This structure is there if you need it. It just have a sane list of fields that most web application are using, you can use it or not, it's really up to you.

Here's the fields of the `tpl.PageData`:

```go
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
```

`Lang` and `Locale` are useful if you want to use the i18n feature.

`CurrentUser` is handy if you want to let your templates know about the current user.

`Env` is useful if your system has multiple environment, like dev, staging, prod and you'd want to do different things based on the env. I personally use if to have a non-minified JavaScript bundle in dev and staging, while a minified one in prod.

`Extra` can be useful for anything that your views need that's not present in the main `Data` field.

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

`Alert` can be use to display flash message to the user, errors and successes etc. The `Notification` structure is this:

```go
type Notification struct {
  Title     template.HTML
  Message   template.HTML
  IsSuccess bool
  IsError   bool
  IsWarning bool
}
```

Usually I have a `web` package with a `render.go` that handles and exposes a `Render` function, here's an example:

```go
package web

import (
	"embed"
	"io"
	"log/slog"
	"net/http"

	"github.com/dstpierre/tpl"
	"github.com/dstpierre/xyz/data/model"
	"github.com/dstpierre/xyz/middleware"
	"golang.org/x/net/xsrftoken"
)

//go:embed all:templates
var fs embed.FS

var templ *tpl.Template

func LoadTemplates() error {
	t, err := tpl.Parse(fs, fmap())
	if err != nil {
		return err
	}

	templ = t

	return nil
}

type BackwardCompatPageData struct {
	tpl.PageData
	Role     model.Roles
	Language string
}

func Render(w io.Writer, r *http.Request, view string, args ...any) {
	d := BackwardCompatPageData{}

	if len(args) > 0 {
		d.Data = args[0]
	}

	if len(args) >= 2 {
		n, ok := args[1].(*tpl.Notification)
		if ok {
			d.Alert = n
		}
	}

	if len(args) >= 3 {
		d.Extra = args[2]
	}

	s, ok := r.Context().Value(middleware.ContextKeySession).(model.Login)
	if ok {
		d.CurrentUser = &s

		d.Role = s.Role
	}

	d.XSRFToken = xsrftoken.Generate(XSRFToken, "", "")
	d.Locale = r.Context().Value(middleware.ContextKeyLanguage).(string)
	d.Lang = d.Locale[:2]

	d.Language = d.Lang

	if err := templ.Render(w, view, d); err != nil {
		slog.Error("error rendering page", "PAGE", view, "ERR", err)
	}
}

```

This is a real-world example, I'm embedding `tpl.PageData` into an existing structure even if it's repeating some field as my existing HTML template were already using `{{ .Language }}` and `tpl.PageData` have a `Lang` field.

> And yes, I was too lazy to replace all `.Language`.

So it's really flexible if you either use it as-is or embed into an existing structure for existing HTML templates.

I'm using it like this in an handler:

```go
func x(w http.ResponseWriter, r *http.Request) {
  flash := &tpl.Notification{Message: "Did not work", isError: true}
  data := actionThatReturnAStruct()
  web.Render(w, r, "app/do.html", data, flash)
}
```

The fact that my `web.Render` functions accept a variadic arguments I'm able to use the function somewhat relative to what happened in the handler. If there's no alert, I only pass the data, if there's no data, I just render the page.

**This is just an **example, you can shape it the way you prefer. This library only facilitate the structuring, parsing, and rendering of templates.

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

## Rendering emails

There's nothing really special regarding emails, other than `tpl` handles their rendering directly, once you have call the `Parse` function you may render any email template like so:

```go
func sendVerifyEmail(token string) error {
  type EmailData struct {
    Link string
  }

  data := EmailData{Link: "https://verify.com/" + token}

  var buf bytes.Buffer
  if err := templ.RenderEmail(&buf, "verify-email.txt", data); err != nil {
    return err
  }

  // you can now send the email and use the bytes as the body
}
```

Your templates in `templates/emails` can access all built-in functions and will also have the same funcmap as your HTML templates.

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

The translation functions expect a language as first argument. This is where the `tpl.PageData` may come handy if you use it directly or embed it in your structure.

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

Translation functions are also exposes, so `tpl.Translate` can be call from your backend if you need translation outside of HTML templates.

## Passing a funcmap

You may have helper functions you'd like to pass to the templates. Here's how:

```go
package main

import (
  "embed"
  "github.com/dstpierre/tpl"
)

//go:embed templates
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

## Built-in functions

`tpl` adds the following functions to the funcmap.

| function | description |
|-----------|------------|
| map | Create a map, useful to pass data to another template  |
| iterate | Allow you to iterate X numbers of time |
| xsrf | Render an hidden input for your XSRF token |
| cut | Remove chars from a string |
| default | Display a fallback if input is nil or zero |
| filesize | Display bytes size in KB, MB, GB, etc |
| slugify | Turn a string into a slug |
| intcomma | Adds , to thousands |
| naturaltime | Display X minutes ago kind of output |

Look at the `testdata/views/app/dashboard.html` for usage examples of these functions.