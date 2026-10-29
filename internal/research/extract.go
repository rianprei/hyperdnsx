package research

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
)

// Body-extraction patterns. Compiled once at package load.
var (
	// jsSrcRe captures JS references in src=/href= attributes and bare .js URLs.
	jsSrcRe = regexp.MustCompile(`(?i)(?:src|href)\s*=\s*["']([^"']+?\.js(?:\?[^"']*)?)["']`)
	jsURLRe = regexp.MustCompile(`(?i)\bhttps?://[^\s"'<>]+?\.js(?:\?[^\s"'<>]*)?`)
	// emailRe is a conservative RFC-ish email matcher.
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
)

// ExtractConfig selects which body-extraction passes run during probing.
type ExtractConfig struct {
	JS      bool
	Emails  bool
	Regexes map[string]*regexp.Regexp // name -> compiled custom regex
}

// Enabled reports whether any extraction pass is configured.
func (c ExtractConfig) Enabled() bool {
	return c.JS || c.Emails || len(c.Regexes) > 0
}

// Extraction holds the artifacts recovered from a response body.
type Extraction struct {
	JSFiles []string
	Emails  []string
	Custom  map[string][]string
}

// Extractor runs the configured body-extraction passes. Recovers the JS-file
// (yabeow), custom-regex (b3r1ch), and email/content (nyxgeek) capabilities
// surveyed from the httpx fork ecosystem.
type Extractor struct {
	cfg ExtractConfig
}

// NewExtractor builds an Extractor. Returns nil when nothing is configured, so
// callers can cheaply skip the stage.
func NewExtractor(cfg ExtractConfig) *Extractor {
	if !cfg.Enabled() {
		return nil
	}
	return &Extractor{cfg: cfg}
}

// CompileRegexes compiles name=pattern specs into a map usable in ExtractConfig.
// A spec without '=' is auto-named rN by its index.
func CompileRegexes(specs []string) (map[string]*regexp.Regexp, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	out := make(map[string]*regexp.Regexp, len(specs))
	for i, spec := range specs {
		name := fmt.Sprintf("r%d", i)
		pattern := spec
		for j := 0; j < len(spec); j++ {
			if spec[j] == '=' {
				name, pattern = spec[:j], spec[j+1:]
				break
			}
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compile regex %q: %w", spec, err)
		}
		out[name] = re
	}
	return out, nil
}

// Extract runs every configured pass over body. baseURL is used to resolve
// relative JS references into absolute URLs. The returned slices are
// deduplicated and sorted for stable output.
func (e *Extractor) Extract(baseURL string, body []byte) Extraction {
	var out Extraction
	text := string(body)

	if e.cfg.JS {
		out.JSFiles = e.extractJS(baseURL, text)
	}
	if e.cfg.Emails {
		out.Emails = sortedUnique(emailRe.FindAllString(text, -1))
	}
	if len(e.cfg.Regexes) > 0 {
		out.Custom = make(map[string][]string, len(e.cfg.Regexes))
		for name, re := range e.cfg.Regexes {
			if matches := customMatches(re, text); len(matches) > 0 {
				out.Custom[name] = matches
			}
		}
		if len(out.Custom) == 0 {
			out.Custom = nil
		}
	}
	return out
}

// extractJS collects JS references from attributes and bare URLs, resolving
// relative paths against baseURL.
func (e *Extractor) extractJS(baseURL, text string) []string {
	var found []string
	base, baseErr := url.Parse(baseURL)
	for _, m := range jsSrcRe.FindAllStringSubmatch(text, -1) {
		ref := m[1]
		if baseErr == nil {
			if abs, err := base.Parse(ref); err == nil {
				ref = abs.String()
			}
		}
		found = append(found, ref)
	}
	found = append(found, jsURLRe.FindAllString(text, -1)...)
	return sortedUnique(found)
}

// customMatches returns submatch group 1 when the pattern has a capture group,
// otherwise the full match. Mirrors the b3r1ch subexpression behavior.
func customMatches(re *regexp.Regexp, text string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(text, -1) {
		if len(m) > 1 && m[1] != "" {
			out = append(out, m[1])
		} else {
			out = append(out, m[0])
		}
	}
	return sortedUnique(out)
}

// sortedUnique deduplicates and sorts a slice, dropping empties.
func sortedUnique(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
