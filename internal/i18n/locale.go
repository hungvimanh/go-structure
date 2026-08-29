package i18n

import "strings"

type Locale string

const (
	LocaleVI Locale = "vi"
	LocaleEN Locale = "en"

	DefaultLocale = LocaleVI
)

var supportedLocales = map[Locale]bool{
	LocaleVI: true,
	LocaleEN: true,
}

func ParseAcceptLanguage(header string) Locale {
	if header == "" {
		return DefaultLocale
	}

	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if tag == "" {
			continue
		}

		lang := strings.ToLower(strings.SplitN(tag, "-", 2)[0])
		if supportedLocales[Locale(lang)] {
			return Locale(lang)
		}
	}

	return DefaultLocale
}
