package research

import (
	"reflect"
	"testing"
)

const sampleBody = `<html><head>
<script src="/static/app.js"></script>
<script src="https://cdn.example.com/lib.js?v=2"></script>
<link href="relative/style.js">
</head><body>
contact us at admin@example.com or sales@example.com
token=ABCD1234 token=WXYZ9999
</body></html>`

func TestExtractJS(t *testing.T) {
	e := NewExtractor(ExtractConfig{JS: true})
	if e == nil {
		t.Fatal("expected non-nil extractor")
	}

	got := e.Extract("https://site.test/index.html", []byte(sampleBody))

	want := []string{
		"https://cdn.example.com/lib.js?v=2",
		"https://site.test/relative/style.js",
		"https://site.test/static/app.js",
	}
	if !reflect.DeepEqual(got.JSFiles, want) {
		t.Errorf("JSFiles = %v, want %v", got.JSFiles, want)
	}
}

func TestExtractEmails(t *testing.T) {
	e := NewExtractor(ExtractConfig{Emails: true})
	got := e.Extract("https://site.test", []byte(sampleBody))
	want := []string{"admin@example.com", "sales@example.com"}
	if !reflect.DeepEqual(got.Emails, want) {
		t.Errorf("Emails = %v, want %v", got.Emails, want)
	}
}

func TestExtractCustomRegexWithCaptureGroup(t *testing.T) {
	// Arrange: capture group => subexpression output (b3r1ch behavior).
	regexes, err := CompileRegexes([]string{`tok=token=(\w+)`})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	e := NewExtractor(ExtractConfig{Regexes: regexes})

	// Act
	got := e.Extract("https://site.test", []byte(sampleBody))

	// Assert: subexpression values, deduped+sorted
	want := []string{"ABCD1234", "WXYZ9999"}
	if !reflect.DeepEqual(got.Custom["tok"], want) {
		t.Errorf("custom[tok] = %v, want %v", got.Custom["tok"], want)
	}
}

func TestExtractCustomRegexNoCaptureGroup(t *testing.T) {
	regexes, err := CompileRegexes([]string{`admin@[a-z.]+`}) // auto-named r0
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	e := NewExtractor(ExtractConfig{Regexes: regexes})
	got := e.Extract("https://site.test", []byte(sampleBody))
	if want := []string{"admin@example.com"}; !reflect.DeepEqual(got.Custom["r0"], want) {
		t.Errorf("custom[r0] = %v, want %v", got.Custom["r0"], want)
	}
}

func TestNewExtractorNilWhenDisabled(t *testing.T) {
	if NewExtractor(ExtractConfig{}) != nil {
		t.Error("expected nil extractor when nothing configured")
	}
}

func TestCompileRegexesError(t *testing.T) {
	if _, err := CompileRegexes([]string{`bad=(`}); err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestCompileRegexesEmpty(t *testing.T) {
	got, err := CompileRegexes(nil)
	if err != nil || got != nil {
		t.Errorf("expected (nil,nil), got (%v,%v)", got, err)
	}
}

func TestSortedUnique(t *testing.T) {
	got := sortedUnique([]string{"b", "a", "b", "", "c", "a"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sortedUnique = %v, want %v", got, want)
	}
	if sortedUnique(nil) != nil {
		t.Error("expected nil for empty input")
	}
}
