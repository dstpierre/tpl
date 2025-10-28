package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/dstpierre/tpl"
)

var keyRegex = regexp.MustCompile(`tp?\s+\.Lang\s+"([^"]+)"`)

var (
	rootPath string
	lang     string
)

func main() {
	flag.StringVar(&rootPath, "path", "", "templates root path")
	flag.StringVar(&lang, "lang", "", "Target language")
	flag.Parse()

	if len(rootPath) == 0 || len(lang) == 0 {
		flag.Usage()
		return
	}

	templateFiles, err := findAllTemplateFiles(rootPath, "*.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(templateFiles) == 0 {
		fmt.Println("No HTML template files found")
		return
	}

	allKeys := make(map[string]struct{})
	for _, file := range templateFiles {
		keys, err := findKeysInFile(file)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", file, err)
			continue
		}
		for key := range keys {
			allKeys[key] = struct{}{}
		}
	}

	msgs, err := parseTargetFile(rootPath, lang)
	if err != nil {
		fmt.Println(err)

	}

	langKeys, err := getTargetKeys(msgs)
	if err != nil {
		fmt.Println(err)
		return
	}

	for key := range allKeys {
		if _, ok := langKeys[key]; !ok {
			msgs = append(msgs, tpl.Text{Key: key})
		}
	}

	if err := saveTargetFile(rootPath, lang, msgs); err != nil {
		fmt.Println(err)
	}
}

func findKeysInFile(filePath string) (map[string]struct{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	keys := make(map[string]struct{})
	matches := keyRegex.FindAllSubmatch(content, -1)

	for _, match := range matches {
		key := string(match[1])
		keys[key] = struct{}{}
	}

	return keys, nil
}

func findAllTemplateFiles(rootPath string, pattern string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		matched, err := filepath.Match(pattern, d.Name())
		if err != nil {
			return err
		}

		if matched {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", rootPath, err)
	}

	return files, nil
}

func parseTargetFile(rootPath, lang string) ([]tpl.Text, error) {
	b, err := os.ReadFile(path.Join(rootPath, "translations", lang+".json"))
	if err != nil {
		return nil, err
	}

	var msgs []tpl.Text
	if err := json.Unmarshal(b, &msgs); err != nil {
		return nil, err
	}

	return msgs, nil
}

func getTargetKeys(msgs []tpl.Text) (map[string]struct{}, error) {
	keys := make(map[string]struct{})
	for _, msg := range msgs {
		keys[msg.Key] = struct{}{}
	}

	return keys, nil
}

func saveTargetFile(rootPath, lang string, msgs []tpl.Text) error {
	b, err := json.MarshalIndent(msgs, "", "\t")
	if err != nil {
		return fmt.Errorf("error converting to JSON: %w", err)
	}

	if err := os.WriteFile(path.Join(rootPath, "translations", lang+".json"), b, 0644); err != nil {
		return fmt.Errorf("error writing target file: %w", err)
	}
	return nil
}
