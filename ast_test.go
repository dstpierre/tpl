package tpl

import (
	"bytes"
	"embed"
	"encoding/gob"
	"testing"
)

//go:embed testdata/*
var fsTest embed.FS

func TestExtractTypeDef(t *testing.T) {
	opt := Option{
		TemplateRootName: "testdata",
	}

	Set(opt)

	var fmap map[string]any = map[string]any{
		"abc": func() string { return "abc" },
	}
	templ, err := Parse(fsTest, fmap)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	writeAST(templ, &buf)

	var m map[string][]string
	r := bytes.NewReader(buf.Bytes())
	if err := gob.NewDecoder(r).Decode(&m); err != nil {
		t.Fatal(err)
	}

	fields, ok := m["app/staticanalysis-err.html"]
	if !ok {
		t.Error("unable to find static analysis html template")
	}

	if len(fields) == 0 {
		t.Fatal("no fields found in tree")
	}

	found := false
	for _, field := range fields {
		t.Log("FIELD:", field)

		if field == `@type:"MyDataType","User",` {
			found = true
			break
		}
	}

	if !found {
		t.Error("could not find type def comment in template")
	}
}
