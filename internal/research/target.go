package research

import (
	"fmt"
	"net/url"
	"strings"
)

// Target is a normalized probe target derived from a raw input line.
type Target struct {
	// Raw is the original input line.
	Raw string
	// Host is the hostname (or IP) to resolve / connect to.
	Host string
	// Port is the explicit port if the input carried one, else "".
	Port string
	// Scheme is the explicit scheme if the input carried one, else "".
	Scheme string
}

// parseTarget normalizes a raw input line into a Target. It accepts bare hosts
// ("example.com"), host:port ("example.com:8443"), and full URLs
// ("https://example.com/path"). Path and query are dropped: the pipeline
// probes origins, not paths. Returns an error for empty/garbage input.
func parseTarget(raw string) (Target, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Target{}, fmt.Errorf("empty target")
	}

	// Full URL with scheme.
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil {
			return Target{}, fmt.Errorf("parse url %q: %w", s, err)
		}
		if u.Hostname() == "" {
			return Target{}, fmt.Errorf("url %q has no host", s)
		}
		return Target{Raw: raw, Host: u.Hostname(), Port: u.Port(), Scheme: u.Scheme}, nil
	}

	// host or host:port
	host := s
	port := ""
	if strings.Count(s, ":") == 1 {
		parts := strings.SplitN(s, ":", 2)
		host, port = parts[0], parts[1]
	}
	if host == "" {
		return Target{}, fmt.Errorf("target %q has no host", s)
	}
	return Target{Raw: raw, Host: host, Port: port}, nil
}

// probeURLs expands a target into the concrete URLs to probe, honoring an
// explicit scheme/port when present and otherwise fanning out over the
// configured default schemes. Each returned entry pairs the URL with its
// scheme and port so the Result can record them without re-parsing.
func (t Target) probeURLs(defaultSchemes []string) []probeURL {
	if t.Scheme != "" {
		return []probeURL{{url: buildURL(t.Scheme, t.Host, t.Port), scheme: t.Scheme, port: effectivePort(t.Scheme, t.Port)}}
	}
	var out []probeURL
	for _, scheme := range defaultSchemes {
		out = append(out, probeURL{
			url:    buildURL(scheme, t.Host, t.Port),
			scheme: scheme,
			port:   effectivePort(scheme, t.Port),
		})
	}
	return out
}

type probeURL struct {
	url    string
	scheme string
	port   string
}

// buildURL assembles a scheme://host[:port] origin URL.
func buildURL(scheme, host, port string) string {
	if port == "" {
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}

// effectivePort returns the explicit port, or the well-known default for the
// scheme when none was supplied.
func effectivePort(scheme, port string) string {
	if port != "" {
		return port
	}
	switch scheme {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}
