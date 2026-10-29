package research

import (
	"hash/fnv"
	"math/bits"
	"strings"
	"sync"
)

// Simhash computes a 64-bit simhash over the token shingles of body. Similar
// documents yield hashes a small Hamming distance apart, which lets the
// pipeline suppress near-duplicate responses (the kno-com fork capability),
// reimplemented self-contained with no external dependency.
func Simhash(body []byte) uint64 {
	shingles := shingle(string(body), 3)
	if len(shingles) == 0 {
		return 0
	}
	var vector [64]int
	for sh, weight := range shingles {
		h := hash64(sh)
		for i := 0; i < 64; i++ {
			if h&(uint64(1)<<uint(i)) != 0 {
				vector[i] += weight
			} else {
				vector[i] -= weight
			}
		}
	}
	var out uint64
	for i := 0; i < 64; i++ {
		if vector[i] > 0 {
			out |= uint64(1) << uint(i)
		}
	}
	return out
}

// HammingDistance returns the number of differing bits between a and b.
func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// shingle tokenizes text into overlapping word n-grams with occurrence counts.
// Weighting by count makes the simhash reflect dominant content.
func shingle(text string, n int) map[string]int {
	fields := strings.Fields(strings.ToLower(text))
	out := make(map[string]int)
	if len(fields) < n {
		for _, f := range fields {
			out[f]++
		}
		return out
	}
	for i := 0; i+n <= len(fields); i++ {
		out[strings.Join(fields[i:i+n], " ")]++
	}
	return out
}

func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// Deduper suppresses near-duplicate results by simhash proximity. It is safe
// for concurrent use.
//
// shortcut: linear scan of seen hashes (fine for typical scan sizes), upgrade
// when scanning very large target sets -> band/LSH index over the hashes.
type Deduper struct {
	threshold int
	mu        sync.Mutex
	seen      []uint64
}

// NewDeduper builds a Deduper. A threshold < 0 is clamped to 0 (exact match).
func NewDeduper(threshold int) *Deduper {
	if threshold < 0 {
		threshold = 0
	}
	return &Deduper{threshold: threshold}
}

// Seen reports whether a result with a hash within threshold bits has already
// been recorded. The first time a (near-)unique hash is presented it is stored
// and Seen returns false; subsequent near-duplicates return true. A zero hash
// (empty body) is never deduplicated.
func (d *Deduper) Seen(hash uint64) bool {
	if hash == 0 {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, h := range d.seen {
		if HammingDistance(h, hash) <= d.threshold {
			return true
		}
	}
	d.seen = append(d.seen, hash)
	return false
}
