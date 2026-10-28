package research

import (
	"testing"
	"time"

	"github.com/projectdiscovery/tlsx/pkg/tlsx/clients"
)

func TestMapTLS(t *testing.T) {
	// Arrange
	before := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	in := &clients.Response{
		Version:  "TLS 1.3",
		Cipher:   "TLS_AES_128_GCM_SHA256",
		Ja3Hash:  "ja3abc",
		Ja3sHash: "ja3sdef",
		JarmHash: "jarmghi",
		CertificateResponse: &clients.CertificateResponse{
			SubjectCN:  "example.com",
			SubjectOrg: []string{"Example Inc"},
			IssuerCN:   "Example CA",
			IssuerOrg:  []string{"Example Trust"},
			SubjectAN:  []string{"example.com", "www.example.com"},
			NotBefore:  before,
			NotAfter:   after,
		},
	}

	// Act
	got := mapTLS(in)

	// Assert
	if got.SubjectCN != "example.com" || got.IssuerCN != "Example CA" {
		t.Errorf("subject/issuer wrong: %+v", got)
	}
	if len(got.SANs) != 2 || got.SANs[0] != "example.com" {
		t.Errorf("SANs wrong: %v", got.SANs)
	}
	if got.JA3Hash != "ja3abc" || got.JARMHash != "jarmghi" {
		t.Errorf("hashes wrong: %+v", got)
	}
	if got.NotBefore != "2026-01-01T00:00:00Z" || got.NotAfter != "2027-01-01T00:00:00Z" {
		t.Errorf("validity wrong: %s / %s", got.NotBefore, got.NotAfter)
	}
}

func TestMapTLSNilCert(t *testing.T) {
	got := mapTLS(&clients.Response{Version: "TLS 1.2"})
	if got.Version != "TLS 1.2" {
		t.Errorf("version lost: %+v", got)
	}
	if got.SubjectCN != "" || len(got.SANs) != 0 {
		t.Errorf("expected empty cert fields, got %+v", got)
	}
}

func TestProbeUnreachableFails(t *testing.T) {
	// Arrange: a port that nothing is listening on, short timeout.
	prober, err := NewProber(ProbeConfig{Timeout: time.Second, RetryMax: 0})
	if err != nil {
		t.Fatalf("NewProber: %v", err)
	}
	base := Result{Input: "127.0.0.1:1", Host: "127.0.0.1"}

	// Act
	got := prober.Probe(base, "http://127.0.0.1:1", "http", "1")

	// Assert
	if !got.Failed || got.Error == "" {
		t.Errorf("expected failed probe, got %+v", got)
	}
	if got.URL != "http://127.0.0.1:1" {
		t.Errorf("URL not carried: %+v", got)
	}
}

func TestFirstToken(t *testing.T) {
	tests := []struct{ in, sep, want string }{
		{"text/html; charset=utf-8", ";", "text/html"},
		{"application/json", ";", "application/json"},
		{"", ";", ""},
	}
	for _, tt := range tests {
		if got := firstToken(tt.in, tt.sep); got != tt.want {
			t.Errorf("firstToken(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
