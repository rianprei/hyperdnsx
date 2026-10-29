package research

import (
	"strings"
	"testing"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
)

func TestSimhashSimilarBodiesClose(t *testing.T) {
	a := Simhash([]byte("the quick brown fox jumps over the lazy dog near the river bank"))
	b := Simhash([]byte("the quick brown fox jumps over the lazy dog near the river side"))
	c := Simhash([]byte("completely unrelated content about astronomy and distant galaxies forming"))

	if a == 0 || b == 0 || c == 0 {
		t.Fatal("simhash of non-empty body must be non-zero")
	}
	near := HammingDistance(a, b)
	far := HammingDistance(a, c)
	if near >= far {
		t.Errorf("expected similar bodies closer than unrelated: near=%d far=%d", near, far)
	}
}

func TestSimhashEmptyBody(t *testing.T) {
	if Simhash(nil) != 0 || Simhash([]byte("")) != 0 {
		t.Error("empty body must hash to 0")
	}
}

func TestDeduperSuppressesNearDuplicates(t *testing.T) {
	// Arrange: realistic near-duplicate pages (same template, tiny change) vs an
	// unrelated page. Simhash is designed for substantial documents, so short
	// strings are not representative — use a page-sized body.
	para := strings.Repeat("the quick brown fox jumps over the lazy dog and then runs across the green field under a bright sky while birds sing softly ", 10)
	base := Simhash([]byte(para + " footer alpha one two three"))
	near := Simhash([]byte(para + " footer alpha one two four")) // tiny change -> small distance
	far := Simhash([]byte(strings.Repeat("astronomy galaxies nebula quasar pulsar cosmic radiation distant starlight dark matter energy ", 10)))

	d := NewDeduper(5)

	// Act + Assert
	if d.Seen(base) {
		t.Fatal("first hash must not be seen")
	}
	if !d.Seen(near) {
		t.Errorf("near-duplicate (distance %d) should be suppressed at threshold 5", HammingDistance(base, near))
	}
	if d.Seen(far) {
		t.Errorf("unrelated body (distance %d) must not be suppressed", HammingDistance(base, far))
	}
}

func TestDeduperZeroHashNeverSuppressed(t *testing.T) {
	d := NewDeduper(5)
	if d.Seen(0) || d.Seen(0) {
		t.Error("zero hash (empty body) must never be treated as duplicate")
	}
}

func TestNewDeduperClampsNegative(t *testing.T) {
	d := NewDeduper(-5)
	if d.threshold != 0 {
		t.Errorf("negative threshold must clamp to 0, got %d", d.threshold)
	}
}

func TestTechAndCPE(t *testing.T) {
	// Arrange: two techs, one with a CPE, one without.
	info := map[string]wappalyzer.AppInfo{
		"Beta":  {CPE: "cpe:/a:x:y"},
		"Alpha": {CPE: ""},
	}

	// Act
	tech, cpe := techAndCPE(info)

	// Assert: tech sorted, CPEs collected non-empty
	if len(tech) != 2 || tech[0] != "Alpha" || tech[1] != "Beta" {
		t.Errorf("tech = %v, want sorted [Alpha Beta]", tech)
	}
	if len(cpe) != 1 || cpe[0] != "cpe:/a:x:y" {
		t.Errorf("cpe = %v, want [cpe:/a:x:y]", cpe)
	}
}

func TestTechAndCPEEmpty(t *testing.T) {
	tech, cpe := techAndCPE(nil)
	if tech != nil || cpe != nil {
		t.Errorf("expected (nil,nil), got (%v,%v)", tech, cpe)
	}
}
