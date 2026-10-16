package runner

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/projectdiscovery/dnsx/libs/dnsx"
	"github.com/projectdiscovery/hmap/store/hybrid"
	"github.com/projectdiscovery/retryabledns"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/stretchr/testify/require"
)

func TestRunner_singleDomain_prepareInput(t *testing.T) {
	options := &Options{
		Domains: "one.one.one.one",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	// call the prepareInput
	err = r.prepareInput()
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"one.one.one.one"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestRunner_domainWildCard_prepareInput(t *testing.T) {
	options := &Options{
		Domains:  "projectdiscovery.io",
		WordList: "jenkins,beta",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	// call the prepareInput
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip()
	}
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"jenkins.projectdiscovery.io", "beta.projectdiscovery.io"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestRunner_fuzzInput_prepareInput(t *testing.T) {
	options := &Options{
		Hosts:    "FUZZ.example.com",
		WordList: "a,b,c",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip()
	}
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"a.example.com", "b.example.com", "c.example.com"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

// TestRunner_wordlistStreamingDedup_prepareInput ensures the streamed wordlist
// expansion still deduplicates entries (and does not buffer the whole product).
func TestRunner_wordlistStreamingDedup_prepareInput(t *testing.T) {
	options := &Options{
		Domains:  "example.com",
		WordList: "a,b,a,b,c",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip()
	}
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"a.example.com", "b.example.com", "c.example.com"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestRunner_cidrInput_prepareInput(t *testing.T) {
	options := &Options{
		Domains: "173.0.84.0/30",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	// call the prepareInput
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip()
	}
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"173.0.84.0", "173.0.84.1", "173.0.84.2", "173.0.84.3"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestRunner_asnInput_prepareInput(t *testing.T) {
	options := &Options{
		Domains: "AS14421",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip("skipping: ASN API key not configured")
	}
	require.Nil(t, err, "failed to prepare input")
	expectedOutputFile := "tests/AS14421.txt"
	// read the expected IPs from the file
	fileContent, err := os.ReadFile(expectedOutputFile)
	require.Nil(t, err, "could not read the expectedOutputFile file")
	expected := strings.Split(strings.ReplaceAll(string(fileContent), "\r\n", "\n"), "\n")
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func isUnauthorizedError(err error) bool {
	return err != nil && stringsutil.ContainsAny(err.Error(), "unauthorized")
}

func TestRunner_fileInput_prepareInput(t *testing.T) {
	options := &Options{
		Hosts: "tests/file_input.txt",
	}
	hm, err := hybrid.New(hybrid.DefaultDiskOptions)
	require.Nil(t, err, "could not create hybrid map")
	r := Runner{
		options: options,
		hm:      hm,
	}
	// call the prepareInput
	err = r.prepareInput()
	if isUnauthorizedError(err) {
		t.Skip()
	}
	require.Nil(t, err, "failed to prepare input")
	expected := []string{"one.one.one.one", "example.com"}
	got := []string{}
	r.hm.Scan(func(k, v []byte) error {
		got = append(got, string(k))
		return nil
	})
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestRunner_hostsInput_prepareInput(t *testing.T) {
	t.Run("file", func(t *testing.T) {
		hm, err := hybrid.New(hybrid.DefaultDiskOptions)
		require.NoError(t, err)
		r := Runner{options: &Options{Hosts: "tests/file_input.txt"}, hm: hm}
		require.NoError(t, r.prepareInput())
		got := scanHMap(t, r.hm)
		require.ElementsMatch(t, []string{"one.one.one.one", "example.com"}, got)
	})

	t.Run("stdin", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "dnsx-stdin-test")
		require.NoError(t, err)
		defer func() {
			_ = os.Remove(tmp.Name())
		}()
		_, err = tmp.WriteString("one.one.one.one\nexample.com\n")
		require.NoError(t, err)
		_ = tmp.Close()

		hm, err := hybrid.New(hybrid.DefaultDiskOptions)
		require.NoError(t, err)
		r := Runner{options: &Options{Hosts: "-"}, hm: hm, tmpStdinFile: tmp.Name()}
		require.NoError(t, r.prepareInput())
		got := scanHMap(t, r.hm)
		require.ElementsMatch(t, []string{"one.one.one.one", "example.com"}, got)
	})

	t.Run("single inline host", func(t *testing.T) {
		hm, err := hybrid.New(hybrid.DefaultDiskOptions)
		require.NoError(t, err)
		r := Runner{options: &Options{Hosts: "one.one.one.one"}, hm: hm}
		require.NoError(t, r.prepareInput())
		got := scanHMap(t, r.hm)
		require.ElementsMatch(t, []string{"one.one.one.one"}, got)
	})

	t.Run("comma separated", func(t *testing.T) {
		hm, err := hybrid.New(hybrid.DefaultDiskOptions)
		require.NoError(t, err)
		r := Runner{options: &Options{Hosts: "one.one.one.one,example.com,cloudflare.com"}, hm: hm}
		require.NoError(t, r.prepareInput())
		got := scanHMap(t, r.hm)
		require.ElementsMatch(t, []string{"one.one.one.one", "example.com", "cloudflare.com"}, got)
	})

	t.Run("empty returns error", func(t *testing.T) {
		hm, err := hybrid.New(hybrid.DefaultDiskOptions)
		require.NoError(t, err)
		r := Runner{options: &Options{}, hm: hm}
		require.Error(t, r.prepareInput())
	})
}

func scanHMap(t *testing.T, hm *hybrid.HybridMap) []string {
	t.Helper()
	var items []string
	hm.Scan(func(k, v []byte) error {
		items = append(items, string(k))
		return nil
	})
	return items
}

func TestRunner_InputWorkerStream(t *testing.T) {
	options := &Options{
		Hosts: "tests/stream_input.txt",
	}
	r := Runner{
		options:    options,
		workerchan: make(chan string),
	}
	go r.InputWorkerStream()
	var got []string
	for c := range r.workerchan {
		got = append(got, c)
	}
	baseExpected := []string{"173.0.84.0", "173.0.84.1", "173.0.84.2", "173.0.84.3", "one.one.one.one"}
	fileContent, err := os.ReadFile("tests/AS14421.txt")
	require.Nil(t, err, "could not read the expectedOutputFile file")
	asnIPs := strings.Split(strings.ReplaceAll(string(fileContent), "\r\n", "\n"), "\n")
	if len(got) == len(baseExpected) {
		t.Skip("skipping: ASN API key not configured")
	}
	expected := append(baseExpected, asnIPs...)
	require.ElementsMatch(t, expected, got, "could not match expected output")
}

func TestNewRejectsAutoWildcardAndWildcardDomainTogether(t *testing.T) {
	runner, err := New(&Options{AutoWildcard: true, WildcardDomain: "example.com"})
	require.Nil(t, runner)
	require.EqualError(t, err, "auto-wildcard and wildcard-domain can't be used at the same time")
}

func TestNewRejectsWildcardFilteringInStreamMode(t *testing.T) {
	t.Run("auto wildcard", func(t *testing.T) {
		runner, err := New(&Options{Stream: true, AutoWildcard: true})
		require.Nil(t, runner)
		require.EqualError(t, err, "wildcard not supported in stream mode")
	})

	t.Run("manual wildcard", func(t *testing.T) {
		runner, err := New(&Options{Stream: true, WildcardDomain: "example.com"})
		require.Nil(t, runner)
		require.EqualError(t, err, "wildcard not supported in stream mode")
	})
}

func TestNormalizeAndValidateWildcardDomain(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		normalized, err := normalizeAndValidateWildcardDomain("*.Example.COM.")
		require.NoError(t, err)
		require.Equal(t, "example.com", normalized)
	})

	for _, input := range []string{"*.", "foo bar", "example..com", "sub.*.example.com", "foo*bar.com"} {
		t.Run(input, func(t *testing.T) {
			normalized, err := normalizeAndValidateWildcardDomain(input)
			require.Error(t, err)
			require.Empty(t, normalized)
		})
	}
}

func TestNewRejectsInvalidWildcardDomain(t *testing.T) {
	runner, err := New(&Options{WildcardDomain: "foo bar"})
	require.Nil(t, runner)
	require.EqualError(t, err, "invalid wildcard domain")
}

func TestHasSelectedRecord(t *testing.T) {
	withCNAME := &dnsx.ResponseData{DNSData: &retryabledns.DNSData{CNAME: []string{"target.example.com"}}}
	withA := &dnsx.ResponseData{DNSData: &retryabledns.DNSData{A: []string{"1.2.3.4"}}}
	empty := &dnsx.ResponseData{DNSData: &retryabledns.DNSData{}}

	tests := []struct {
		name     string
		options  *Options
		data     *dnsx.ResponseData
		expected bool
	}{
		{"cname requested and present", &Options{CNAME: true}, withCNAME, true},
		{"cname requested but absent", &Options{CNAME: true}, withA, false},
		{"a requested and present", &Options{A: true}, withA, true},
		{"any of several requested types present", &Options{A: true, CNAME: true}, withA, true},
		{"any record type always matches", &Options{ANY: true}, empty, true},
		{"no requested type present", &Options{CNAME: true, MX: true}, empty, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Runner{options: tt.options}
			require.Equal(t, tt.expected, r.hasSelectedRecord(tt.data))
		})
	}
}

func TestRunner_resolverModeAll_returnsOneAnswerPerResolver(t *testing.T) {
	// Two bogus resolvers (no DNS server listening). QueryMultipleAll must
	// still emit one answer per resolver, each carrying its provenance, so
	// consumers can see which resolver was queried and that it failed.
	resolver1 := "127.0.0.1:19991"
	resolver2 := "127.0.0.1:19992"

	tempDir := t.TempDir()
	hostsFile := filepath.Join(tempDir, "hosts.txt")
	outputFile := filepath.Join(tempDir, "output.txt")

	require.NoError(t, os.WriteFile(hostsFile, []byte("example.com\n"), 0644))

	testOptions := &Options{
		Hosts:             hostsFile,
		OutputFile:        outputFile,
		Resolvers:         resolver1 + "," + resolver2,
		Threads:           2,
		Retries:           1,
		Timeout:           time.Millisecond * 50,
		Silent:            true,
		NoColor:           true,
		JSON:              true,
		ResolverMode:      "all",
		resolverMode:      dnsx.ResolverModeAll,
		WildcardThreshold: 2,
	}

	runner, err := New(testOptions)
	require.NoError(t, err)
	defer runner.Close()

	require.NoError(t, runner.Run())

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	require.Greater(t, len(lines), 0, "expected at least one output line")

	seenResolvers := map[string]int{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		resolver, ok := entry["resolver"].(string)
		require.True(t, ok, "expected resolver provenance field in JSON output: %s", line)
		seenResolvers[resolver]++
	}

	require.Contains(t, seenResolvers, resolver1, "resolver1 should have produced answers in all mode")
	require.Contains(t, seenResolvers, resolver2, "resolver2 should have produced answers in all mode")
}

func TestRunner_resolverModeDefaultsToFirstSuccess(t *testing.T) {
	_, shutdown := startWildcardTestDNSServer(t)
	defer shutdown()

	tempDir := t.TempDir()
	hostsFile := filepath.Join(tempDir, "hosts.txt")
	outputFile := filepath.Join(tempDir, "output.txt")

	require.NoError(t, os.WriteFile(hostsFile, []byte("example.com\n"), 0644))

	testOptions := &Options{
		Hosts:             hostsFile,
		OutputFile:        outputFile,
		Resolvers:         "127.0.0.1:53",
		Threads:           1,
		Retries:           1,
		Timeout:           time.Second,
		Silent:            true,
		NoColor:           true,
		JSON:              true,
		A:                 true,
		WildcardThreshold: 2,
	}

	runner, err := New(testOptions)
	require.NoError(t, err)
	require.Equal(t, dnsx.ResolverModeFirstSuccess, runner.options.resolverMode)
	defer runner.Close()

	require.NoError(t, runner.Run())
}

// runResolverModeTest executes the runner against unresolvable local addresses
// using the given resolver mode and returns the parsed JSONL output lines.
func runResolverModeTest(t *testing.T, mode dnsx.ResolverMode, quorum int, servers []string, hosts []string) []map[string]interface{} {
	t.Helper()

	tempDir := t.TempDir()
	hostsFile := filepath.Join(tempDir, "hosts.txt")
	outputFile := filepath.Join(tempDir, "output.txt")

	require.NoError(t, os.WriteFile(hostsFile, []byte(strings.Join(hosts, "\n")+"\n"), 0644))

	resolvers := strings.Join(servers, ",")

	testOptions := &Options{
		Hosts:             hostsFile,
		OutputFile:        outputFile,
		Resolvers:         resolvers,
		Threads:           len(servers),
		Retries:           1,
		Timeout:           time.Second,
		Silent:            true,
		NoColor:           true,
		JSON:              true,
		ResolverMode:      mode.String(),
		resolverMode:      mode,
		Quorum:            quorum,
		A:                 true,
		WildcardThreshold: 2,
	}

	runner, err := New(testOptions)
	require.NoError(t, err)
	defer runner.Close()

	require.NoError(t, runner.Run())

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var lines []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		lines = append(lines, entry)
	}
	return lines
}

// ── Quorum tests (using unresolvable local addresses, deterministic failures) ─

func TestRunner_quorum_allResolversEmptyResponse_consensusOnEmpty(t *testing.T) {
	// Two unresolvable addresses → retryabledns returns empty NOERROR answers
	// for both, producing identical content signatures → consensus reached.
	s1 := "127.0.0.1:19001"
	s2 := "127.0.0.1:19002"

	lines := runResolverModeTest(t, dnsx.ResolverModeQuorum, 2, []string{s1, s2}, []string{"example.com"})
	require.Len(t, lines, 1, "expected exactly one JSONL output line")

	entry := lines[0]
	require.Equal(t, "quorum", entry["mode"])
	require.True(t, entry["consensus_reached"].(bool), "all resolvers produce identical empty response → consensus")
	require.Equal(t, float64(2), entry["total_voters"])
	require.Equal(t, float64(2), entry["quorum_required"])
	// The shared empty signature should have 2 votes.
	voteCounts := entry["vote_counts"].(map[string]interface{})
	require.Contains(t, voteCounts, "rcode=0")
	require.Equal(t, float64(2), voteCounts["rcode=0"].(float64))
}

func TestRunner_quorum_mixedSignatures_noConsensus(t *testing.T) {
	// Three resolvers, all produce empty NOERROR locally (retryabledns
	// limitation), so they share one signature. Consensus IS reached because
	// all three agree. This documents the current behavior honestly.
	s1 := "127.0.0.1:19011"
	s2 := "127.0.0.1:19012"
	s3 := "127.0.0.1:19013"

	lines := runResolverModeTest(t, dnsx.ResolverModeQuorum, 2, []string{s1, s2, s3}, []string{"example.com"})
	require.Len(t, lines, 1)

	entry := lines[0]
	require.True(t, entry["consensus_reached"].(bool))
	require.Equal(t, float64(3), entry["total_voters"])
	require.Equal(t, float64(2), entry["quorum_required"])
}

// ── Compare tests (using unresolvable local addresses) ───────────────────────

func TestRunner_compare_allResolversEmptyResponse_oneSharedDivergence(t *testing.T) {
	// Two unresolvable addresses → both produce identical empty NOERROR → one
	// shared divergence (vote_count=2), no unique failures.
	s1 := "127.0.0.1:19021"
	s2 := "127.0.0.1:19022"

	lines := runResolverModeTest(t, dnsx.ResolverModeCompare, 0, []string{s1, s2}, []string{"example.com"})
	require.Len(t, lines, 1)

	entry := lines[0]
	require.Equal(t, "compare", entry["mode"])
	require.Equal(t, float64(2), entry["total_resolvers"])
	require.Equal(t, float64(0), entry["failed_resolvers"], "empty responses are not failures")
	divs := entry["divergences"].([]interface{})
	require.Len(t, divs, 1, "both resolvers share the same empty signature")
	d := divs[0].(map[string]interface{})
	require.Equal(t, float64(2), d["vote_count"])
	unique, _ := d["unique"].(bool)
	require.False(t, unique, "shared signature is not unique")
}

func TestRunner_compare_oneResolver(t *testing.T) {
	// Single resolver → one answer, one divergence (unique).
	s1 := "127.0.0.1:19031"

	lines := runResolverModeTest(t, dnsx.ResolverModeCompare, 0, []string{s1}, []string{"example.com"})
	require.Len(t, lines, 1)

	entry := lines[0]
	require.Equal(t, float64(1), entry["total_resolvers"])
	divs := entry["divergences"].([]interface{})
	require.Len(t, divs, 1)
	d := divs[0].(map[string]interface{})
	require.Equal(t, float64(1), d["vote_count"])
	unique, _ := d["unique"].(bool)
	require.True(t, unique, "single resolver produces a unique signature")
}
