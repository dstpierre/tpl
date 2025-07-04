package tpl

import (
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func enhanceFuncMap(fmap map[string]any) {
	addTranslationFunctions(fmap)
	addInternationalizationFunctions(fmap)
	addHelperFunctions(fmap)
	addHumanizeFunctions(fmap)
}

func addTranslationFunctions(fmap map[string]any) {
	fmap["t"] = Translate
	fmap["tp"] = TranslatePlural
	fmap["tf"] = TranslateFormat
	fmap["tfp"] = TranslateFormatPlural
}

func addInternationalizationFunctions(fmap map[string]any) {
	fmap["shortdate"] = ToDate
	fmap["currency"] = ToCurrency
}

func addHelperFunctions(fmap map[string]any) {
	fmap["map"] = func(v ...any) map[string]any {
		if len(v)%2 != 0 {
			panic("call to map should have a key and value of even pairs")
		}

		m := make(map[string]any)
		for i := 0; i < len(v); i += 2 {
			key, ok := v[i].(string)
			if !ok {
				panic(fmt.Sprintf("key for the map function should be string: %v", v[i]))
			}

			m[key] = v[i+1]
		}

		return m
	}

	fmap["iterate"] = func(max uint) []uint {
		l := make([]uint, max)
		var idx uint
		for idx = 0; idx < max; idx++ {
			l[idx] = idx
		}
		return l
	}

	fmap["xsrf"] = func(token string) template.HTML {
		return template.HTML(
			fmt.Sprintf(`<input type="hidden" name="xsrf-token" value="%s">`, token),
		)
	}

	fmap["cut"] = func(v, s string) string {
		return strings.Replace(s, v, "", -1)
	}

	fmap["default"] = func(fallback, value any) any {
		if value == nil {
			return fallback
		}

		// Use reflect to check for zero values of various types
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.String:
			if val.Len() == 0 {
				return fallback
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if val.Int() == 0 {
				return fallback
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if val.Uint() == 0 {
				return fallback
			}
		case reflect.Float32, reflect.Float64:
			if val.Float() == 0.0 {
				return fallback
			}
		case reflect.Bool:
			if !val.Bool() {
				return fallback
			}
		case reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
			if !val.IsValid() || val.IsNil() || (val.Kind() == reflect.Slice && val.Len() == 0) || (val.Kind() == reflect.Map && val.Len() == 0) {
				return fallback
			}
		default:
			// For any other type, if it's not nil, consider it "set"
		}

		return value
	}

	fmap["filesize"] = func(bytes float64) string {
		const (
			KB = 1024.0
			MB = 1024.0 * KB
			GB = 1024.0 * MB
			TB = 1024.0 * GB
			PB = 1024.0 * TB
		)

		switch {
		case bytes < KB:
			return fmt.Sprintf("%.0f B", bytes)
		case bytes < MB:
			return fmt.Sprintf("%.1f KB", bytes/KB)
		case bytes < GB:
			return fmt.Sprintf("%.1f MB", bytes/MB)
		case bytes < TB:
			return fmt.Sprintf("%.1f GB", bytes/GB)
		case bytes < PB:
			return fmt.Sprintf("%.1f TB", bytes/TB)
		default:
			return fmt.Sprintf("%.1f PB", bytes/PB)
		}
	}

	nonWordCharOrSpace := regexp.MustCompile(`[^\w\s-]`)
	multipleHyphens := regexp.MustCompile(`-+`)
	whitespace := regexp.MustCompile(`\s+`)

	fmap["slugify"] = func(s string) string {
		s = strings.ToLower(s)

		s = whitespace.ReplaceAllString(s, "-")
		s = nonWordCharOrSpace.ReplaceAllString(s, "")
		s = multipleHyphens.ReplaceAllString(s, "-")
		s = strings.Trim(s, "-")

		return s
	}

}

func addHumanizeFunctions(fmap map[string]any) {
	fmap["intcomma"] = func(i int64) string {
		s := fmt.Sprintf("%d", i)
		n := len(s)
		if n <= 3 {
			return s
		}

		// Calculate the position of the first comma
		firstComma := n % 3
		if firstComma == 0 {
			firstComma = 3
		}

		var result strings.Builder
		result.WriteString(s[:firstComma])

		for j := firstComma; j < n; j += 3 {
			result.WriteString(",")
			result.WriteString(s[j : j+3])
		}

		return result.String()
	}

	fmap["naturaltime"] = func(t time.Time) string {
		now := time.Now()
		diff := now.Sub(t)

		// Handle future dates (from now)
		if diff < 0 {
			diff = -diff
			if diff < 1*time.Minute {
				return "in " + formatDuration(diff)
			} else if diff < 1*time.Hour {
				minutes := int(diff.Minutes())
				return fmt.Sprintf("in %d minute%s", minutes, plural(minutes))
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("in %d hour%s", hours, plural(hours))
			} else if diff < 30*24*time.Hour { // Roughly 30 days for a month
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("in %d day%s", days, plural(days))
			} else if diff < 365*24*time.Hour { // Roughly 365 days for a year
				months := int(diff.Hours() / (30 * 24)) // Approximate month
				return fmt.Sprintf("in %d month%s", months, plural(months))
			} else {
				years := int(diff.Hours() / (365 * 24)) // Approximate year
				return fmt.Sprintf("in %d year%s", years, plural(years))
			}
		}

		// Handle past dates (ago)
		if diff < 1*time.Minute {
			seconds := int(diff.Seconds())
			if seconds < 10 {
				return "just now"
			}
			return fmt.Sprintf("%d second%s ago", seconds, plural(seconds))
		} else if diff < 1*time.Hour {
			minutes := int(diff.Minutes())
			return fmt.Sprintf("%d minute%s ago", minutes, plural(minutes))
		} else if diff < 24*time.Hour {
			hours := int(diff.Hours())
			return fmt.Sprintf("%d hour%s ago", hours, plural(hours))
		} else if diff < 30*24*time.Hour { // Roughly 30 days for a month
			days := int(diff.Hours() / 24)
			return fmt.Sprintf("%d day%s ago", days, plural(days))
		} else if diff < 365*24*time.Hour { // Roughly 365 days for a year
			months := int(diff.Hours() / (30 * 24)) // Approximate month
			return fmt.Sprintf("%d month%s ago", months, plural(months))
		} else {
			years := int(diff.Hours() / (365 * 24)) // Approximate year
			return fmt.Sprintf("%d year%s ago", years, plural(years))
		}
	}
}

func formatDuration(d time.Duration) string {
	if d < 1*time.Minute {
		seconds := int(d.Seconds())
		if seconds < 10 {
			return "just now"
		}
		return fmt.Sprintf("%d second%s", seconds, plural(seconds))
	}
	return ""
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
