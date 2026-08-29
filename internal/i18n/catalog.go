package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Catalog struct {
	messages map[string]map[Locale]string
}

func LoadCatalog(root string) (*Catalog, error) {
	messages := make(map[string]map[Locale]string)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".i18n.json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var file map[string]map[string]string
		if err := json.Unmarshal(data, &file); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		for code, translations := range file {
			if _, exists := messages[code]; exists {
				return fmt.Errorf("duplicate i18n code %q found in %s", code, path)
			}

			byLocale := make(map[Locale]string, len(translations))
			for lang, tmpl := range translations {
				byLocale[Locale(lang)] = tmpl
			}
			messages[code] = byLocale
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Catalog{messages: messages}, nil
}
