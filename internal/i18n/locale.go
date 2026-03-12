package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kittors/AgentFlow/internal/config"
)

type Locale string

const (
	LocaleZH Locale = "zh"
	LocaleEN Locale = "en"
)

type Catalog struct {
	lang string
}

func NewCatalog() Catalog {
	return Catalog{lang: DetectLocale()}
}

func NewCatalogWithLanguage(lang string) Catalog {
	if locale, ok := normalizeLocale(lang); ok {
		return Catalog{lang: string(locale)}
	}
	return NewCatalog()
}

func DetectLocale() string {
	if locale, ok := LoadPreferredLocale(); ok {
		return locale
	}
	return DetectLocaleFromEnvironment()
}

func DetectLocaleFromEnvironment() string {
	if locale, ok := normalizeLocale(os.Getenv("AGENTFLOW_LANG")); ok {
		return string(locale)
	}

	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		if locale, ok := normalizeLocale(os.Getenv(key)); ok {
			return string(locale)
		}
	}

	if runtime.GOOS == "windows" {
		if locale, ok := normalizeLocale(os.Getenv("LANG")); ok {
			return string(locale)
		}
	}

	return string(LocaleEN)
}

func LoadPreferredLocale() (string, bool) {
	path, err := preferredLocalePath()
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	if locale, ok := normalizeLocale(string(data)); ok {
		return string(locale), true
	}
	return "", false
}

func SavePreferredLocale(lang string) error {
	locale, ok := normalizeLocale(lang)
	if !ok {
		return fmt.Errorf("unsupported locale: %s", lang)
	}
	path, err := preferredLocalePath()
	if err != nil {
		return err
	}
	return config.SafeWrite(path, []byte(string(locale)+"\n"), 0o644)
}

func preferredLocalePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".agentflow", "preferences", "locale"), nil
}

func (c Catalog) Language() string {
	return c.lang
}

func (c Catalog) Msg(zh, en string) string {
	if c.lang == string(LocaleZH) {
		return zh
	}
	return en
}

func Msg(zh, en string) string {
	return NewCatalog().Msg(zh, en)
}

func normalizeLocale(input string) (Locale, bool) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch {
	case normalized == "":
		return "", false
	case strings.HasPrefix(normalized, "zh") || normalized == "cn" || normalized == "chinese":
		return LocaleZH, true
	case strings.HasPrefix(normalized, "en") || normalized == "english":
		return LocaleEN, true
	default:
		return "", false
	}
}
