package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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

func TestCheckIgnoresMalformedCachedVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.0.3-main.abcdef1"}`))
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")
	checker.Now = func() time.Time {
		return time.Unix(1_700_000_000, 0)
	}

	if err := checker.writeCache(cacheEntry{Latest: "continuous", Timestamp: checker.Now().Unix()}); err != nil {
		t.Fatalf("writeCache returned error: %v", err)
	}

	originalAPI := releaseAPIOverride
	releaseAPIOverride = server.URL
	defer func() { releaseAPIOverride = originalAPI }()

	result, err := checker.Check("1.0.3", Options{CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.Latest != "1.0.3-main.abcdef1" {
		t.Fatalf("expected malformed cache entry to be ignored, got %#v", result)
	}
}

func TestCheckRefreshesStaleCachedMainBuild(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.0.3-main.d4c9af6"}`))
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")
	checker.Now = func() time.Time {
		return time.Unix(1_700_000_000, 0)
	}

	if err := checker.writeCache(cacheEntry{Latest: "1.0.3-main.dfa4830", Timestamp: checker.Now().Unix()}); err != nil {
		t.Fatalf("writeCache returned error: %v", err)
	}

	originalAPI := releaseAPIOverride
	releaseAPIOverride = server.URL
	defer func() { releaseAPIOverride = originalAPI }()

	result, err := checker.Check("1.0.3-main.d4c9af6", Options{CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.UpdateAvailable {
		t.Fatalf("expected stale cached main build to be refreshed, got %#v", result)
	}
	if result.Latest != "1.0.3-main.d4c9af6" {
		t.Fatalf("expected latest version to refresh from API, got %#v", result)
	}
}

func TestCheckDoesNotOfferDowngrade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.0.1-main.deadbee"}`))
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")

	originalAPI := releaseAPIOverride
	releaseAPIOverride = server.URL
	defer func() { releaseAPIOverride = originalAPI }()

	result, err := checker.Check("1.0.2", Options{Force: true, CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.UpdateAvailable {
		t.Fatalf("expected no downgrade suggestion, got %#v", result)
	}
}

func TestCheckPrefersContinuousReleaseEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/tag":
			_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.0.3-main.abcdef1"}`))
		case "/latest":
			_, _ = writer.Write([]byte(`{"tag_name":"v1.0.3","name":"v1.0.3"}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")

	originalTagAPI := releaseTagAPI
	originalLatestAPI := releaseLatestAPI
	originalOverride := releaseAPIOverride
	releaseTagAPI = server.URL + "/tag"
	releaseLatestAPI = server.URL + "/latest"
	releaseAPIOverride = ""
	defer func() {
		releaseTagAPI = originalTagAPI
		releaseLatestAPI = originalLatestAPI
		releaseAPIOverride = originalOverride
	}()

	result, err := checker.Check("1.0.3", Options{Force: true, CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result.UpdateAvailable || result.Latest != "1.0.3-main.abcdef1" {
		t.Fatalf("expected continuous release metadata to win, got %#v", result)
	}
}

func TestCheckFallsBackToLatestWhenContinuousEndpointUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/tag":
			http.NotFound(writer, request)
		case "/latest":
			_, _ = writer.Write([]byte(`{"tag_name":"v1.2.3"}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")

	originalTagAPI := releaseTagAPI
	originalLatestAPI := releaseLatestAPI
	originalOverride := releaseAPIOverride
	releaseTagAPI = server.URL + "/tag"
	releaseLatestAPI = server.URL + "/latest"
	releaseAPIOverride = ""
	defer func() {
		releaseTagAPI = originalTagAPI
		releaseLatestAPI = originalLatestAPI
		releaseAPIOverride = originalOverride
	}()

	result, err := checker.Check("1.0.0", Options{Force: true, CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result.UpdateAvailable || result.Latest != "1.2.3" {
		t.Fatalf("expected fallback to latest release metadata, got %#v", result)
	}
}

func TestCheckPrefersReleaseNameForContinuousRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.0.3-main.abcdef1"}`))
	}))
	defer server.Close()

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")

	originalAPI := releaseAPIOverride
	releaseAPIOverride = server.URL
	defer func() { releaseAPIOverride = originalAPI }()

	result, err := checker.Check("1.0.3", Options{Force: true, CacheTTLHours: 72})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result.UpdateAvailable {
		t.Fatalf("expected continuous release to be treated as an update, got %#v", result)
	}
	if result.Latest != "1.0.3-main.abcdef1" {
		t.Fatalf("expected release name to be used as latest version, got %#v", result)
	}
}

func TestShouldUpdateHandlesEqualCoreVersionWithDifferentMainBuild(t *testing.T) {
	if !shouldUpdate("1.0.3-main.aaaaaaa", "1.0.3-main.bbbbbbb") {
		t.Fatal("expected different main builds with same core version to offer an update")
	}
	if shouldUpdate("1.1.0", "1.0.3-main.bbbbbbb") {
		t.Fatal("expected newer local stable version not to downgrade to older continuous release")
	}
}

func TestAssetNameForPlatform(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{goos: "darwin", goarch: "arm64", want: "agentflow-darwin-arm64"},
		{goos: "linux", goarch: "amd64", want: "agentflow-linux-amd64"},
		{goos: "windows", goarch: "amd64", want: "agentflow-windows-amd64.exe"},
	}

	for _, test := range tests {
		got, err := assetNameForPlatform(test.goos, test.goarch)
		if err != nil {
			t.Fatalf("assetNameForPlatform(%q, %q) returned error: %v", test.goos, test.goarch, err)
		}
		if got != test.want {
			t.Fatalf("assetNameForPlatform(%q, %q) = %q, want %q", test.goos, test.goarch, got, test.want)
		}
	}
}

func TestSelfUpdateReplacesExecutableOnUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update replacement test is Unix-only")
	}

	assetName, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("assetNameForPlatform returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/latest":
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.2.3-main.abcdef1","assets":[{"name":"` + assetName + `","browser_download_url":"http://` + request.Host + `/download"}]}`))
		case "/download":
			_, _ = writer.Write([]byte("new-binary"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	originalAPI := releaseAPIOverride
	originalExecutablePath := executablePath
	originalEvalSymlinks := evalSymlinks
	t.Cleanup(func() {
		releaseAPIOverride = originalAPI
		executablePath = originalExecutablePath
		evalSymlinks = originalEvalSymlinks
	})

	executable := filepath.Join(t.TempDir(), "agentflow")
	if err := os.WriteFile(executable, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	releaseAPIOverride = server.URL + "/latest"
	executablePath = func() (string, error) { return executable, nil }
	evalSymlinks = func(path string) (string, error) { return path, nil }

	checker := NewChecker()
	checker.Client = server.Client()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")
	checker.Now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	result, err := checker.SelfUpdate("1.0.0")
	if err != nil {
		t.Fatalf("SelfUpdate returned error: %v", err)
	}
	if !result.UpdateAvailable || result.Latest != "1.2.3-main.abcdef1" {
		t.Fatalf("unexpected result: %#v", result)
	}

	content, err := os.ReadFile(executable)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "new-binary" {
		t.Fatalf("expected executable to be replaced, got %q", string(content))
	}
}

func TestSelfUpdateAllowsSlowDownloadWithDefaultCheckerTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update replacement test is Unix-only")
	}

	assetName, err := assetNameForPlatform(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("assetNameForPlatform returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/latest":
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"tag_name":"continuous","name":"1.2.4-main.abcdef2","assets":[{"name":"` + assetName + `","browser_download_url":"http://` + request.Host + `/download"}]}`))
		case "/download":
			time.Sleep(6 * time.Second)
			_, _ = writer.Write([]byte("slow-binary"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	originalAPI := releaseAPIOverride
	originalExecutablePath := executablePath
	originalEvalSymlinks := evalSymlinks
	t.Cleanup(func() {
		releaseAPIOverride = originalAPI
		executablePath = originalExecutablePath
		evalSymlinks = originalEvalSymlinks
	})

	executable := filepath.Join(t.TempDir(), "agentflow")
	if err := os.WriteFile(executable, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	releaseAPIOverride = server.URL + "/latest"
	executablePath = func() (string, error) { return executable, nil }
	evalSymlinks = func(path string) (string, error) { return path, nil }

	checker := NewChecker()
	checker.CacheFile = filepath.Join(t.TempDir(), "version_cache.json")
	checker.Now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	result, err := checker.SelfUpdate("1.0.0")
	if err != nil {
		t.Fatalf("SelfUpdate returned error after slow download: %v", err)
	}
	if !result.UpdateAvailable || result.Latest != "1.2.4-main.abcdef2" {
		t.Fatalf("unexpected result: %#v", result)
	}

	content, err := os.ReadFile(executable)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "slow-binary" {
		t.Fatalf("expected executable to be replaced after slow download, got %q", string(content))
	}
}
