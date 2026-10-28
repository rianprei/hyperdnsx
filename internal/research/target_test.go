package research

import (
	"reflect"
	"testing"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Target
		wantErr bool
	}{
		{
			name: "bare host",
			raw:  "example.com",
			want: Target{Raw: "example.com", Host: "example.com"},
		},
		{
			name: "host with port",
			raw:  "example.com:8443",
			want: Target{Raw: "example.com:8443", Host: "example.com", Port: "8443"},
		},
		{
			name: "full https url drops path",
			raw:  "https://example.com/admin?x=1",
			want: Target{Raw: "https://example.com/admin?x=1", Host: "example.com", Scheme: "https"},
		},
		{
			name: "url with explicit port",
			raw:  "http://example.com:8080/",
			want: Target{Raw: "http://example.com:8080/", Host: "example.com", Port: "8080", Scheme: "http"},
		},
		{
			name:    "empty input",
			raw:     "   ",
			wantErr: true,
		},
		{
			name:    "scheme without host",
			raw:     "https://",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTarget(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTarget(%q) err = %v, wantErr = %v", tt.raw, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTarget(%q) = %+v, want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestProbeURLs(t *testing.T) {
	tests := []struct {
		name    string
		target  Target
		schemes []string
		want    []probeURL
	}{
		{
			name:    "bare host fans out over default schemes",
			target:  Target{Host: "example.com"},
			schemes: []string{"https", "http"},
			want: []probeURL{
				{url: "https://example.com", scheme: "https", port: "443"},
				{url: "http://example.com", scheme: "http", port: "80"},
			},
		},
		{
			name:    "explicit scheme is honored",
			target:  Target{Host: "example.com", Scheme: "https"},
			schemes: []string{"https", "http"},
			want:    []probeURL{{url: "https://example.com", scheme: "https", port: "443"}},
		},
		{
			name:    "explicit port carried through",
			target:  Target{Host: "example.com", Port: "8443"},
			schemes: []string{"https"},
			want:    []probeURL{{url: "https://example.com:8443", scheme: "https", port: "8443"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.target.probeURLs(tt.schemes)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("probeURLs = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEffectivePort(t *testing.T) {
	tests := []struct {
		scheme, port, want string
	}{
		{"https", "", "443"},
		{"http", "", "80"},
		{"https", "9000", "9000"},
		{"ftp", "", ""},
	}
	for _, tt := range tests {
		if got := effectivePort(tt.scheme, tt.port); got != tt.want {
			t.Errorf("effectivePort(%q,%q) = %q, want %q", tt.scheme, tt.port, got, tt.want)
		}
	}
}

func TestAppendUnique(t *testing.T) {
	// Arrange
	start := []string{"1.1.1.1"}

	// Act
	got := appendUnique(start, "1.1.1.1", "2.2.2.2", "", "2.2.2.2", "3.3.3.3")

	// Assert
	want := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("appendUnique = %v, want %v", got, want)
	}
}
