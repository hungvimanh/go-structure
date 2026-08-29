package i18n

import (
	"fmt"

	"category-service/internal/shared/validation"
)

func (c *Catalog) Translate(locale Locale, fe validation.FieldError) string {
	tmpl, ok := c.lookup(locale, fe.Code)
	if !ok {
		return fe.Code
	}
	if len(fe.Params) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, fe.Params...)
}

func (c *Catalog) TranslateAll(locale Locale, errs validation.Errors) map[string]string {
	out := make(map[string]string, len(errs))
	for _, fe := range errs {
		out[fe.Field] = c.Translate(locale, fe)
	}
	return out
}

func (c *Catalog) lookup(locale Locale, code string) (string, bool) {
	byLocale, ok := c.messages[code]
	if !ok {
		return "", false
	}
	if tmpl, ok := byLocale[locale]; ok {
		return tmpl, true
	}
	if tmpl, ok := byLocale[DefaultLocale]; ok {
		return tmpl, true
	}
	return "", false
}
