package tpl_test

import (
	"bytes"
	"embed"
	"strings"
	"testing"
	"time"

	"github.com/dstpierre/tpl"
)

//go:embed testdata/*
var fsTest embed.FS

var fmap map[string]any = map[string]any{
	"abc": func() string {
		return "from custom func map"
	},
}

func load(t *testing.T) *tpl.Template {
	opts := tpl.Option{TemplateRootName: "testdata"}
	tpl.Set(opts)

	templ, err := tpl.Parse(fsTest, fmap)
	if err != nil {
		t.Fatal(err)
	}

	return templ
}

type pagedata struct {
	Text   string
	Date   time.Time
	Amount float64
}

func render(t *testing.T, templ *tpl.Template, view string) string {
	data := tpl.PageData{
		Lang:   "fr",
		Locale: "fr-CA",
		Title:  "unit-test",
		Data:   pagedata{Text: "unit-test", Date: time.Now(), Amount: 1234.56},
	}
	var buf bytes.Buffer
	if err := templ.Render(&buf, view, data); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

func TestLoadTemplates(t *testing.T) {
	load(t)
}

func TestRender(t *testing.T) {
	templ := load(t)

	body := render(t, templ, "layout/user-login.html")
	if !strings.Contains(body, "<p>unit-test</p>") {
		t.Errorf("body does not contains unit-test: %s", body)
	}
}

func TestAppLayoutNav(t *testing.T) {
	templ := load(t)

	body := render(t, templ, "app/dashboard.html")
	if !strings.Contains(body, "<p>Main nav here</p>") {
		t.Errorf("can't find main nav in body: %s", body)
	} else if !strings.Contains(body, "func map") {
		t.Errorf("can't find func map in body: %s", body)
	}
}
