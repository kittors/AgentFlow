package i18n

import (
	"os"
	"runtime"
	"strings"
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

func DetectLocale() string {
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
