package update

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckUsesCacheAndTrimsVersionPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")
	checker.Now = func() time.Time {
		return time.Unix(1_700_000_000, 0)
	}

	originalAPI := releaseAPIOverride
	releaseAPIOverride = server.URL
	defer func() { releaseAPIOverride = originalAPI }()

	result, err := checker.Check("1.0.0", Options{CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result.UpdateAvailable || result.Latest != "1.2.3" {
		t.Fatalf("unexpected result: %#v", result)
	}

	releaseAPIOverride = "http://127.0.0.1:9"
	cached, err := checker.Check("1.0.0", Options{CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("expected cached result, got error: %v", err)
	}
	if cached.Latest != "1.2.3" {
		t.Fatalf("expected cached latest version, got %#v", cached)
	}
}
