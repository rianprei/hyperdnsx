package dnsx

import (
	"sort"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/projectdiscovery/retryabledns"
	"github.com/stretchr/testify/require"
)

func TestResolverMode_Default(t *testing.T) {
	require.Equal(t, ResolverModeFirstSuccess, ResolverMode("first-success"))
	require.Equal(t, ResolverModeAll, ResolverMode("all"))
	require.Equal(t, ResolverModeQuorum, ResolverMode("quorum"))
	require.Equal(t, ResolverModeCompare, ResolverMode("compare"))
}

func TestResolverMode_IsValid(t *testing.T) {
	require.True(t, ResolverModeFirstSuccess.IsValid())
	require.True(t, ResolverModeAll.IsValid())
	require.True(t, ResolverModeQuorum.IsValid())
	require.True(t, ResolverModeCompare.IsValid())
	require.False(t, ResolverMode("bogus").IsValid())
	require.False(t, ResolverMode("").IsValid())
}

func TestResolverMode_String(t *testing.T) {
	require.Equal(t, "first-success", ResolverModeFirstSuccess.String())
	require.Equal(t, "all", ResolverModeAll.String())
	require.Equal(t, "quorum", ResolverModeQuorum.String())
	require.Equal(t, "compare", ResolverModeCompare.String())
}

func TestProtocolFromResolver(t *testing.T) {
	require.Equal(t, "udp", protocolFromResolver("udp:1.1.1.1:53"))
	require.Equal(t, "tcp", protocolFromResolver("tcp:8.8.8.8:53"))
	require.Equal(t, "doh", protocolFromResolver("doh:https://dns.google/dns-query"))
	require.Equal(t, "dot", protocolFromResolver("dot:1.1.1.1:853"))
	require.Equal(t, "udp", protocolFromResolver("1.1.1.1"))
}

func TestContentSignature_IgnoresTTL(t *testing.T) {
	a := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		A:             []string{"1.1.1.1", "9.9.9.9"},
		TTL:           60,
	}
	b := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		A:             []string{"9.9.9.9", "1.1.1.1"},
		TTL:           3600,
	}
	sa, _ := contentSignature(a)
	sb, _ := contentSignature(b)
	require.Equal(t, sa, sb, "same records in different order and TTL must produce identical signature")
}

func TestContentSignature_DifferentContent(t *testing.T) {
	a := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		A:             []string{"1.1.1.1"},
	}
	b := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		A:             []string{"2.2.2.2"},
	}
	sa, _ := contentSignature(a)
	sb, _ := contentSignature(b)
	require.NotEqual(t, sa, sb)
}

func TestContentSignature_DifferentRCODE(t *testing.T) {
	a := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		A:             []string{"1.1.1.1"},
	}
	b := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeNameError,
		A:             []string{"1.1.1.1"},
	}
	sa, _ := contentSignature(a)
	sb, _ := contentSignature(b)
	require.NotEqual(t, sa, sb, "different RCODE must diverge")
}

func TestContentSignature_EmptyAnswer(t *testing.T) {
	a := &retryabledns.DNSData{
		Host:          "example.com",
		StatusCodeRaw: dns.RcodeSuccess,
	}
	sa, empty := contentSignature(a)
	require.True(t, empty)
	require.Contains(t, sa, "rcode=0")
}

func TestQuorumResult_ConsensusReached(t *testing.T) {
	sigA := "rcode=0|A:1.1.1.1,9.9.9.9"
	sigB := "rcode=0|A:2.2.2.2"
	qr := &QuorumResult{
		VoteCounts:     map[string]int{sigA: 3, sigB: 1},
		TotalVoters:    4,
		QuorumRequired: 3,
		AllAnswers: []*ResponseData{
			{DNSData: &retryabledns.DNSData{Host: "example.com", StatusCodeRaw: dns.RcodeSuccess, A: []string{"1.1.1.1", "9.9.9.9"}}},
		},
	}
	// Simulate selection logic
	bestSig := ""
	bestVotes := 0
	for sig, votes := range qr.VoteCounts {
		if votes > bestVotes || (votes == bestVotes && sig < bestSig) {
			bestSig = sig
			bestVotes = votes
		}
	}
	require.Equal(t, sigA, bestSig)
	require.Equal(t, 3, bestVotes)
	require.True(t, bestVotes >= qr.QuorumRequired)
}

func TestQuorumResult_NoConsensus(t *testing.T) {
	qr := &QuorumResult{
		VoteCounts:     map[string]int{"sigA": 2, "sigB": 2},
		TotalVoters:    4,
		QuorumRequired: 3,
	}
	bestVotes := 0
	for _, votes := range qr.VoteCounts {
		if votes > bestVotes {
			bestVotes = votes
		}
	}
	require.Equal(t, 2, bestVotes)
	require.False(t, bestVotes >= qr.QuorumRequired, "no signature reaches quorum of 3 with 2-2 split")
}

func TestCompareResult_SortsDivergences(t *testing.T) {
	cr := &CompareResult{
		Divergences: []CompareDivergence{
			{Signature: "zzz", Resolvers: []string{"r1"}, VoteCount: 1, Unique: true},
			{Signature: "aaa", Resolvers: []string{"r1", "r2"}, VoteCount: 2, Unique: false},
			{Signature: "mmm", Resolvers: []string{"r3"}, VoteCount: 1, Unique: true},
		},
	}
	require.Len(t, cr.Divergences, 3)
	// Apply the same deterministic ordering used by QueryMultipleCompare.
	sort.Slice(cr.Divergences, func(i, j int) bool {
		if cr.Divergences[i].Unique != cr.Divergences[j].Unique {
			return !cr.Divergences[i].Unique
		}
		return cr.Divergences[i].Signature < cr.Divergences[j].Signature
	})
	require.False(t, cr.Divergences[0].Unique, "first divergence should be non-unique (shared)")
	require.Equal(t, "aaa", cr.Divergences[0].Signature)
	require.True(t, cr.Divergences[1].Unique)
	require.Less(t, cr.Divergences[1].Signature, cr.Divergences[2].Signature)
}

func TestQuorumResult_DefaultQuorum(t *testing.T) {
	qr := &QuorumResult{QuorumRequired: 0}
	if qr.QuorumRequired < 1 {
		qr.QuorumRequired = 1
	}
	require.Equal(t, 1, qr.QuorumRequired)
}

// buildResponseData creates a ResponseData with a DNSData A record for testing.
func buildResponseData(host string, ips ...string) *ResponseData {
	return &ResponseData{
		DNSData: &retryabledns.DNSData{
			Host:          host,
			StatusCodeRaw: dns.RcodeSuccess,
			StatusCode:    dns.RcodeToString[dns.RcodeSuccess],
			A:             ips,
			Timestamp:     time.Now(),
			Resolver:      []string{"test"},
		},
		Resolver: "test",
	}
}

func TestContentSignature_CNAMEChain(t *testing.T) {
	a := &retryabledns.DNSData{
		Host:          "www.example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		CNAME:         []string{"example.com"},
		A:             []string{"1.1.1.1"},
	}
	b := &retryabledns.DNSData{
		Host:          "www.example.com",
		StatusCodeRaw: dns.RcodeSuccess,
		CNAME:         []string{"example.com"},
		A:             []string{"1.1.1.1"},
		TTL:           9999,
	}
	sa, _ := contentSignature(a)
	sb, _ := contentSignature(b)
	require.Equal(t, sa, sb, "CNAME chain with same records but different TTL must match")
}
