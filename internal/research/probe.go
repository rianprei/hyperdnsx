package research

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/projectdiscovery/httpx/common/httpx"
	"github.com/projectdiscovery/tlsx/pkg/tlsx/clients"
)

// ProbeConfig controls the HTTP probe stage. Zero values fall back to httpx
// defaults, so callers only set what they want to override.
type ProbeConfig struct {
	Timeout         time.Duration
	RetryMax        int
	FollowRedirects bool
	MaxRedirects    int
	TLSGrab         bool
	RandomAgent     bool
	Threads         int
	Unsafe          bool
	CustomHeaders   map[string][]string
	// TLSImpersonate selects a ClientHello profile (e.g. "chrome") when set,
	// recovering the browser-impersonation capability surveyed from the
	// rabb1t-dev fork. Empty means the default Go ClientHello.
	TLSImpersonate string
	// Extractor, when non-nil, runs body-extraction passes on each response.
	Extractor *Extractor
}

// Prober wraps the httpx probing engine. It is safe for concurrent use: the
// underlying *httpx.HTTPX performs per-request work and holds no mutable
// per-call state.
type Prober struct {
	engine    *httpx.HTTPX
	tlsGrab   bool
	extractor *Extractor
}

// NewProber builds a Prober from cfg, layering overrides onto httpx defaults.
func NewProber(cfg ProbeConfig) (*Prober, error) {
	opts := httpx.DefaultOptions
	if cfg.Timeout > 0 {
		opts.Timeout = cfg.Timeout
	}
	if cfg.RetryMax > 0 {
		opts.RetryMax = cfg.RetryMax
	}
	if cfg.Threads > 0 {
		opts.Threads = cfg.Threads
	}
	if cfg.MaxRedirects > 0 {
		opts.MaxRedirects = cfg.MaxRedirects
	}
	opts.FollowRedirects = cfg.FollowRedirects
	opts.TLSGrab = cfg.TLSGrab
	opts.RandomAgent = cfg.RandomAgent
	opts.Unsafe = cfg.Unsafe
	opts.TlsImpersonate = cfg.TLSImpersonate
	if len(cfg.CustomHeaders) > 0 {
		opts.CustomHeaders = cfg.CustomHeaders
	}

	engine, err := httpx.New(&opts)
	if err != nil {
		return nil, fmt.Errorf("init httpx engine: %w", err)
	}
	return &Prober{engine: engine, tlsGrab: cfg.TLSGrab, extractor: cfg.Extractor}, nil
}

// Probe performs a single HTTP request against targetURL and returns a copy of
// base enriched with the HTTP fingerprint. base carries the DNS stage; it is
// never mutated (the method takes it by value).
func (p *Prober) Probe(base Result, targetURL, scheme, port string) Result {
	out := base.cloneWithHTTP(targetURL, scheme, port)

	req, err := p.engine.NewRequest(http.MethodGet, targetURL)
	if err != nil {
		out.Failed = true
		out.Error = fmt.Sprintf("build request: %v", err)
		return out
	}

	resp, err := p.engine.Do(req, httpx.UnsafeOptions{})
	if err != nil {
		out.Failed = true
		out.Error = err.Error()
		return out
	}

	out.StatusCode = resp.StatusCode
	out.ContentLength = resp.ContentLength
	out.Words = resp.Words
	out.Lines = resp.Lines
	out.HTTP2 = resp.HTTP2
	out.ResponseTime = resp.Duration.String()
	out.Title = httpx.ExtractTitle(resp)
	out.WebServer = normalizeHeader(resp.GetHeader("Server"))
	out.ContentType = firstToken(normalizeHeader(resp.GetHeader("Content-Type")), ";")
	out.Location = normalizeHeader(resp.GetHeader("Location"))

	if p.tlsGrab && resp.TLSData != nil {
		out.TLS = mapTLS(resp.TLSData)
	}
	if p.extractor != nil && len(resp.Data) > 0 {
		ex := p.extractor.Extract(targetURL, resp.Data)
		out.JSFiles = ex.JSFiles
		out.Emails = ex.Emails
		out.Extract = ex.Custom
	}
	return out
}

// mapTLS projects the tlsx client response into the compact TLSInfo record.
func mapTLS(d *clients.Response) *TLSInfo {
	info := &TLSInfo{
		Version:  d.Version,
		Cipher:   d.Cipher,
		JA3Hash:  d.Ja3Hash,
		JA3SHash: d.Ja3sHash,
		JARMHash: d.JarmHash,
	}
	if d.CertificateResponse != nil {
		c := d.CertificateResponse
		info.SubjectCN = c.SubjectCN
		info.SubjectOrg = c.SubjectOrg
		info.IssuerCN = c.IssuerCN
		info.IssuerOrg = c.IssuerOrg
		info.SANs = c.SubjectAN
		if !c.NotBefore.IsZero() {
			info.NotBefore = c.NotBefore.Format(time.RFC3339)
		}
		if !c.NotAfter.IsZero() {
			info.NotAfter = c.NotAfter.Format(time.RFC3339)
		}
	}
	return info
}

// firstToken returns the substring before the first occurrence of sep, trimmed.
func firstToken(s, sep string) string {
	if i := strings.Index(s, sep); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
