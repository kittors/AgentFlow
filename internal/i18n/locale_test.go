package i18n

import "testing"

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   Locale
		wantOK bool
	}{
		{name: "empty", input: "", want: "", wantOK: false},
		{name: "zh_cn", input: "zh_CN.UTF-8", want: LocaleZH, wantOK: true},
		{name: "english", input: "english", want: LocaleEN, wantOK: true},
		{name: "unknown", input: "fr_FR", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := normalizeLocale(tt.input)
			if got != tt.want || gotOK != tt.wantOK {
				t.Fatalf("normalizeLocale(%q) = (%q, %v), want (%q, %v)", tt.input, got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}
