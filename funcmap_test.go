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

func TestBuiltIns(t *testing.T) {
	templ := load(t)
	body := render(t, templ, "app/dashboard.html")
	if !strings.Contains(body, `<input type="hidden" name="xsrf-token" value="xsrf-token-here">`) {
		t.Error("cannot find XSRF token input")
	}

	if !strings.Contains(body, "oktest123") {
		t.Error("cut does not work")
	}

	if !strings.Contains(body, "no user set") {
		t.Error("default does not work")
	}

	if !strings.Contains(body, "1.2 KB") {
		t.Error("filesize is not working")
	}

	if !strings.Contains(body, "a-title-that-do-have") {
		t.Error("slugify does not work")
	}

}

func TestHumanize(t *testing.T) {
	templ := load(t)
	body := render(t, templ, "app/dashboard.html")

	if !strings.Contains(body, "12,321") {
		t.Error("intcomma does not work")
	}

	if !strings.Contains(body, "5 minutes ago") {
		t.Error("naturaltime does not work")
	}
}
