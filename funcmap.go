package tpl

import "fmt"

func enhanceFuncMap(fmap map[string]any) {
	addTranslationFunctions(fmap)
	addInternationalizationFunctions(fmap)
	addHelperFunctions(fmap)
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
}
