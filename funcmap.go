package tpl

func enhanceFuncMap(fmap map[string]any) {
	addTranslationFunctions(fmap)
	addInternationalizationFunctions(fmap)
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
