package tpl_test

import (
	"strings"
	"testing"
	"time"
)

func TestTranslationFunctions(t *testing.T) {
	templ := load(t)
	body := render(t, templ, "app/i18n.html")
	if !strings.Contains(body, "<h1>Allo tout le monde</h1>") {
		t.Errorf("can't find hello-world transaltion: %s", body)
	} else if !strings.Contains(body, "<p>Bonjour personnes</p>") {
		t.Errorf("plural value not found in %s", body)
	}
}

func TestInternationalization(t *testing.T) {
	templ := load(t)
	body := render(t, templ, "app/i18n.html")

	nowInCA := time.Now().Format("02-01-2006")
	if !strings.Contains(body, "<em>"+nowInCA+"</em>") {
		t.Errorf("can't find Canadian date formatted: %s", body)
	} else if !strings.Contains(body, "<em>1234.56 $</em>") {
		t.Errorf("can't find Canadian currency formatted: %s", body)
	}
}
