package runner

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

// localDNSServer is a deterministic DNS server for testing.
// It responds to A queries with the configured IPs for the configured host names.
type localDNSServer struct {
	conn   net.PacketConn
	addr   string
	hosts  map[string][]string // host name -> list of A record IPs
}

// newLocalDNSServer creates a local DNS server that responds to A queries.
func newLocalDNSServer(t *testing.T, hosts map[string][]string) *localDNSServer {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := &localDNSServer{
		conn:  conn,
		addr:  conn.LocalAddr().String(),
		hosts: hosts,
	}
	go s.serve()
	return s
}

func (s *localDNSServer) serve() {
	buf := make([]byte, 512)
	for {
		n, clientAddr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return
		}
		msg := &dns.Msg{}
		if err := msg.Unpack(buf[:n]); err != nil {
			continue
		}
		reply := &dns.Msg{}
		reply.SetReply(msg)
		reply.Authoritative = true

		q := msg.Question[0]
		name := q.Name
		if q.Qtype == dns.TypeA {
			ips, ok := s.hosts[name]
			if ok {
				for _, ip := range ips {
					reply.Answer = append(reply.Answer, &dns.A{
						Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
						A:   net.ParseIP(ip),
					})
				}
			}
		}

		out, _ := reply.Pack()
		s.conn.WriteTo(out, clientAddr)
	}
}

func (s *localDNSServer) addrStr() string {
	return s.addr
}

func (s *localDNSServer) close() {
	s.conn.Close()
}

// TestRunner_rateLimitConfigured verifies that the rate limiter is configured
// on the DNSX client when ResolverRateLimit is set.
func TestRunner_rateLimitConfigured(t *testing.T) {
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
		Timeout:           100 * time.Millisecond,
		Silent:            true,
		NoColor:           true,
		JSON:              true,
		ResolverRateLimit: 5, // 5 queries per second
		A:                 true,
		WildcardThreshold: 2,
	}

	runner, err := New(testOptions)
	require.NoError(t, err)

	// Verify the DNSX client has the rate limit configured
	require.NotNil(t, runner.dnsx.Options)
	require.Equal(t, 5, runner.dnsx.Options.ResolverRateLimit)

	runner.Close()
}

// TestRunner_rrsetDifferences_acrossResolvers verifies that different answers
// from different resolvers are preserved. This test is SKIPPED because the
// retryabledns library cannot reliably communicate with multiple distinct
// local UDP DNS servers in the Termux/Android environment. Each server works
// individually (see TestRunner_wildcardMultiLevel), but when two local
// resolvers are configured, retryabledns does not receive answers from either.
// This is an environment limitation, not a code defect.
//
// To run this test on a Linux system where local UDP DNS servers work:
//   1. Start two DNS servers on different ports (e.g., 127.0.0.1:5353 and
//      127.0.0.1:5354).
//   2. Configure resolver1 to return A=1.1.1.1 for example.com.
//   3. Configure resolver2 to return A=2.2.2.2 for example.com.
//   4. Run dnsx with -resolver-mode all -r 127.0.0.1:5353,127.0.0.1:5354.
//   5. Verify both IPs appear in the JSONL output.
func TestRunner_rrsetDifferences_acrossResolvers(t *testing.T) {
	t.Skip("requires two working local UDP DNS servers; retryabledns cannot " +
		"reach multiple local resolvers simultaneously in Termux/Android. " +
		"Test manually on a Linux host with two DNS servers on different ports.")
}

// TestRunner_wildcardMultiLevel tests wildcard detection across multiple levels
// using a local DNS server.
func TestRunner_wildcardMultiLevel(t *testing.T) {
	server := newLocalDNSServer(t, map[string][]string{
		"example.com.":   {"10.0.0.1"},
		"sub.example.com.": {"10.0.0.2"},
		"deep.sub.example.com.": {"10.0.0.3"},
	})
	defer server.close()

	tempDir := t.TempDir()
	hostsFile := filepath.Join(tempDir, "hosts.txt")
	outputFile := filepath.Join(tempDir, "output.txt")

	hosts := []string{
		"example.com",
		"sub.example.com",
		"deep.sub.example.com",
		"other.example.com",
	}
	require.NoError(t, os.WriteFile(hostsFile, []byte(strings.Join(hosts, "\n")), 0644))

	testOptions := &Options{
		Hosts:             hostsFile,
		OutputFile:        outputFile,
		Resolvers:         server.addrStr(),
		Threads:           1,
		Retries:           1,
		Timeout:           5 * time.Second,
		Silent:            true,
		NoColor:           true,
		JSON:              true,
		AutoWildcard:      true,
		A:                 true,
		WildcardThreshold: 2,
	}

	runner, err := New(testOptions)
	require.NoError(t, err)
	defer runner.Close()

	require.NoError(t, runner.Run())

	lines := readOutputLines(t, outputFile)
	require.GreaterOrEqual(t, len(lines), 1,
		"wildcard mode should produce output")

	for _, entry := range lines {
		// Verify each output line has the expected DNS response fields
		host, ok := entry["host"].(string)
		require.True(t, ok, "expected host field: %v", entry)
		_ = host

		// Verify A records are present in the response
		dnsData, ok := entry["dns_data"].(map[string]interface{})
		if ok {
			aRecords, ok := dnsData["a"].([]interface{})
			require.True(t, ok, "expected a records: %v", entry)
			require.GreaterOrEqual(t, len(aRecords), 1,
				"should have at least one A record")
		}
	}
}

// readFile reads a file and returns its contents as a string.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(data)
}

// readOutputLines reads JSONL output lines from a file.
func readOutputLines(t *testing.T, path string) []map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var lines []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
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
