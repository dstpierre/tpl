package tpl

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
)

type Text struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	PluralValue string `json:"plural"`
}

var messages map[string]Text

func loadTranslations(fs embed.FS) error {
	messages = make(map[string]Text)

	files, err := load(fs, config.TemplateRootName, "translations")
	if err != nil {
		slog.Warn("loading translation files", "ERR", err)
		return nil
	}

	for _, file := range files {
		var msgs []Text
		b, err := fs.ReadFile(file.fullPath)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, &msgs); err != nil {
			return err
		}

		fillTranslations(file.name, msgs)
	}

	return nil
}

func fillTranslations(name string, msgs []Text) {
	lang := strings.TrimSuffix(name, filepath.Ext(name))

	for _, msg := range msgs {
		key := fmt.Sprintf("%s_%s", lang, msg.Key)
		messages[key] = msg
	}
}

// GetMessageFromKey returns the Text structure for a giving language and key.
func GetMessageFromKey(lang, key string) Text {
	k := fmt.Sprintf("%s_%s", lang, key)

	v, ok := messages[k]
	if !ok {
		return Text{Key: key, Value: "not found"}
	}

	return v
}

// Translate returns the proper value based on language and key.
func Translate(lang, key string) string {
	return GetMessageFromKey(lang, key).Value
}

// TranslatePlural returns the proper version based on language, key, and number
func TranslatePlural(lang, key string, num int64) string {
	msg := GetMessageFromKey(lang, key)
	if num > 1 && len(msg.PluralValue) > 0 {
		return msg.PluralValue
	}
	return msg.Value
}

// TranslateFormat returns the formatted text based on language and key
func TranslateFormat(lang, key string, values []any) string {
	return fmt.Sprintf(GetMessageFromKey(lang, key).Value, values...)
}

// TranslateFormatPlural returns the proper formatted text based on language,
// key, and number.
func TranslateFormatPlural(lang, key string, num int64, values []any) string {
	s := TranslatePlural(lang, key, num)
	return fmt.Sprintf(s, values...)
}
