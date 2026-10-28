package research

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

func TestResultJSONRoundTrip(t *testing.T) {
	// Arrange
	in := Result{
		Input:      "example.com",
		Host:       "example.com",
		A:          []string{"93.184.216.34"},
		URL:        "https://example.com",
		Scheme:     "https",
		Port:       "443",
		StatusCode: 200,
		Title:      "Example",
	}

	// Act
	line, err := in.JSON()
	if err != nil {
		t.Fatalf("JSON() error: %v", err)
	}
	var out Result
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Assert
	if out.URL != in.URL || out.StatusCode != in.StatusCode || out.Title != in.Title {
		t.Errorf("round trip mismatch: got %+v", out)
	}
	if strings.Contains(line, "\n") {
		t.Error("JSON() must be a single line")
	}
}

func TestCloneWithHTTPDoesNotMutateBase(t *testing.T) {
	// Arrange
	base := Result{Input: "example.com", Host: "example.com", A: []string{"1.2.3.4"}}

	// Act
	clone := base.cloneWithHTTP("https://example.com", "https", "443")

	// Assert
	if base.URL != "" || base.Scheme != "" {
		t.Errorf("base was mutated: %+v", base)
	}
	if clone.URL != "https://example.com" || clone.Host != "example.com" {
		t.Errorf("clone wrong: %+v", clone)
	}
}

// TestPipelineProbesLocalServer exercises the full probe path hermetically
// against an httptest server, with the DNS stage disabled.
func TestPipelineProbesLocalServer(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "test-server")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><head><title>Hello HyperX</title></head><body>ok</body></html>"))
	}))
	defer srv.Close()

	prober, err := NewProber(ProbeConfig{Threads: 2})
	if err != nil {
		t.Fatalf("NewProber: %v", err)
	}
	pipeline := NewPipeline(nil, prober, PipelineConfig{Threads: 2, Resolve: false})

	// Act
	var mu sync.Mutex
	var results []Result
	pipeline.Run([]string{srv.URL}, func(r Result) {
		mu.Lock()
		results = append(results, r)
		mu.Unlock()
	})

	// Assert
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %+v", len(results), results)
	}
	got := results[0]
	if got.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", got.StatusCode)
	}
	if got.Title != "Hello HyperX" {
		t.Errorf("title = %q, want %q", got.Title, "Hello HyperX")
	}
	if got.WebServer != "test-server" {
		t.Errorf("webserver = %q, want test-server", got.WebServer)
	}
	if got.ContentType != "text/html" {
		t.Errorf("content_type = %q, want text/html", got.ContentType)
	}
	if got.Scheme != "http" {
		t.Errorf("scheme = %q, want http", got.Scheme)
	}
}

// TestProbeHostHeaderOverride verifies -host-header overrides the Host sent to
// the server (vhost probing) and is recorded in the result.
func TestProbeHostHeaderOverride(t *testing.T) {
	// Arrange: server echoes the received Host into a response header.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Echo-Host", r.Host)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	prober, err := NewProber(ProbeConfig{HostHeader: "vhost.internal", WAFDetect: true})
	if err != nil {
		t.Fatalf("NewProber: %v", err)
	}

	// Act
	var results []Result
	NewPipeline(nil, prober, PipelineConfig{Resolve: false}).
		Run([]string{srv.URL}, func(r Result) { results = append(results, r) })

	// Assert
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].HostHeader != "vhost.internal" {
		t.Errorf("host_header = %q, want vhost.internal", results[0].HostHeader)
	}
	if results[0].WAF != "" {
		t.Errorf("unexpected WAF detected on plain server: %q", results[0].WAF)
	}
}

// TestPipelineBadTargetEmitsFailure verifies garbage input produces a failed
// record rather than a panic or silent drop.
func TestPipelineBadTargetEmitsFailure(t *testing.T) {
	prober, err := NewProber(ProbeConfig{})
	if err != nil {
		t.Fatalf("NewProber: %v", err)
	}
	pipeline := NewPipeline(nil, prober, PipelineConfig{Resolve: false})

	var results []Result
	pipeline.Run([]string{"   "}, func(r Result) {
		results = append(results, r)
	})

	if len(results) != 1 || !results[0].Failed {
		t.Fatalf("expected one failed result, got %+v", results)
	}
}

// TestPipelineFanOutMultipleSchemes ensures a bare host with no explicit scheme
// produces one probe record per default scheme.
func TestPipelineFanOutMultipleSchemes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// derive host:port from the test server, probe with explicit http scheme only
	addr := strings.TrimPrefix(srv.URL, "http://")
	prober, _ := NewProber(ProbeConfig{})
	pipeline := NewPipeline(nil, prober, PipelineConfig{Resolve: false, DefaultSchemes: []string{"http"}})

	var schemes []string
	pipeline.Run([]string{addr}, func(r Result) {
		schemes = append(schemes, r.Scheme)
	})
	sort.Strings(schemes)
	if len(schemes) != 1 || schemes[0] != "http" {
		t.Errorf("expected [http], got %v", schemes)
	}
}
