// Package research implements the HyperDNSx deep-research pipeline: it fuses
// the DNS resolution engine (libs/dnsx) with the httpx HTTP probing engine
// into a single resolve -> probe -> enrich flow that emits one unified record
// per probed endpoint.
package research

import (
	"encoding/json"
	"strings"
)

// Result is the unified deep-research record. It carries the DNS provenance
// recovered during the resolve stage together with the HTTP fingerprint
// recovered during the probe stage, so a single JSONL line fully describes an
// endpoint: what it resolved to and what it served.
type Result struct {
	// Input is the original target as provided by the caller.
	Input string `json:"input"`

	// --- DNS stage ---
	Host        string   `json:"host,omitempty"`
	A           []string `json:"a,omitempty"`
	AAAA        []string `json:"aaaa,omitempty"`
	CNAME       []string `json:"cname,omitempty"`
	DNSResolver string   `json:"dns_resolver,omitempty"`

	// --- HTTP stage ---
	URL           string `json:"url,omitempty"`
	Scheme        string `json:"scheme,omitempty"`
	Port          string `json:"port,omitempty"`
	StatusCode    int    `json:"status_code,omitempty"`
	Title         string `json:"title,omitempty"`
	WebServer     string `json:"webserver,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	ContentLength int    `json:"content_length,omitempty"`
	Words         int    `json:"words,omitempty"`
	Lines         int    `json:"lines,omitempty"`
	Location      string `json:"location,omitempty"`
	HTTP2         bool   `json:"http2,omitempty"`
	ResponseTime  string `json:"response_time,omitempty"`

	// --- TLS (populated when -tls-grab and the endpoint is HTTPS) ---
	TLS *TLSInfo `json:"tls,omitempty"`

	// --- status ---
	Failed bool   `json:"failed,omitempty"`
	Error  string `json:"error,omitempty"`
}

// TLSInfo is a compact projection of the certificate data recovered during the
// TLS handshake. It intentionally mirrors only the fields useful for research
// correlation (subject, issuer, SANs) and omits the raw certificate.
type TLSInfo struct {
	SubjectCN  string   `json:"subject_cn,omitempty"`
	SubjectOrg []string `json:"subject_org,omitempty"`
	IssuerCN   string   `json:"issuer_cn,omitempty"`
	IssuerOrg  []string `json:"issuer_org,omitempty"`
	SANs       []string `json:"sans,omitempty"`
	NotBefore  string   `json:"not_before,omitempty"`
	NotAfter   string   `json:"not_after,omitempty"`
	Version    string   `json:"version,omitempty"`
	Cipher     string   `json:"cipher,omitempty"`
	JA3Hash    string   `json:"ja3_hash,omitempty"`
	JA3SHash   string   `json:"ja3s_hash,omitempty"`
	JARMHash   string   `json:"jarm_hash,omitempty"`
}

// JSON renders the result as a single compact JSON line (no trailing newline).
func (r *Result) JSON() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// cloneWithHTTP returns a copy of the DNS-stage result with HTTP fields reset,
// so one resolved host can fan out into several per-URL records without the
// probe stage mutating the shared DNS record (immutable-by-copy).
func (r Result) cloneWithHTTP(url, scheme, port string) Result {
	r.URL = url
	r.Scheme = scheme
	r.Port = port
	return r
}

// normalizeHeader trims and collapses a header value for stable output.
func normalizeHeader(v string) string {
	return strings.TrimSpace(v)
}
