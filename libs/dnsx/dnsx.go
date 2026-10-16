package dnsx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	miekgdns "github.com/miekg/dns"
	"github.com/projectdiscovery/cdncheck"
	"github.com/projectdiscovery/ratelimit"
	retryabledns "github.com/projectdiscovery/retryabledns"
	iputil "github.com/projectdiscovery/utils/ip"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

// DNSX is structure to perform dns lookups
type DNSX struct {
	dnsClient *retryabledns.Client
	Options   *Options
	cdn       *cdncheck.Client
	// resolverLimiters holds a per-resolver rate limiter when
	// Options.ResolverRateLimit is positive.
	resolverLimiters map[string]*ratelimit.Limiter
}

// Options contains configuration options
type Options struct {
	BaseResolvers     []string
	MaxRetries        int
	QuestionTypes     []uint16
	Trace             bool
	TraceMaxRecursion int
	Hostsfile         bool
	OutputCDN         bool
	QueryAll          bool
	Proxy             string
	Timeout           time.Duration
	// ResolverMode controls how queries are dispatched to resolvers.
	// Defaults to first-success for compatibility with upstream dnsx.
	ResolverMode ResolverMode
	// ResolverRateLimit limits DNS requests per second per resolver
	// (disabled when <= 0).
	ResolverRateLimit int
}

// ResponseData to show output result
type ResponseData struct {
	*retryabledns.DNSData
	// Provenance fields recorded per answer so consumers can trace which
	// resolver produced which answer, over which transport, and how long it took.
	Resolver  string       `json:"resolver,omitempty" csv:"resolver"`
	Protocol  string       `json:"protocol,omitempty" csv:"protocol"`
	Latency   string       `json:"latency,omitempty" csv:"latency"`
	Attempts  int          `json:"attempts,omitempty" csv:"attempts"`
	IsCDNIP   bool         `json:"cdn,omitempty" csv:"cdn"`
	CDNName   string       `json:"cdn-name,omitempty" csv:"cdn-name"`
	CDNType   string       `json:"cdn-type,omitempty" csv:"cdn-type"`
	ASN       *AsnResponse `json:"asn,omitempty" csv:"asn"`
	QueryTime string       `json:"query-time,omitempty" csv:"query-time"`
}
type AsnResponse struct {
	AsNumber  string   `json:"as-number,omitempty" csv:"as_number"`
	AsName    string   `json:"as-name,omitempty" csv:"as_name"`
	AsCountry string   `json:"as-country,omitempty" csv:"as_country"`
	AsRange   []string `json:"as-range,omitempty" csv:"as_range"`
}

func (o *AsnResponse) String() string {
	return fmt.Sprintf("[%v, %v, %v]", o.AsNumber, o.AsName, o.AsCountry)
}

// ResolverMode defines how queries are dispatched to the configured resolvers.
type ResolverMode string

const (
	// ResolverModeFirstSuccess queries resolvers until the first valid answer
	// (historical dnsx behavior, kept as default for compatibility).
	ResolverModeFirstSuccess ResolverMode = "first-success"
	// ResolverModeAll queries every configured resolver and preserves each
	// answer with its provenance (resolver, protocol, latency, attempts, RCODE).
	ResolverModeAll ResolverMode = "all"
	// ResolverModeQuorum classifies answers by consensus across resolvers.
	ResolverModeQuorum ResolverMode = "quorum"
	// ResolverModeCompare shows every answer without picking a winner,
	// flagging divergences between resolvers.
	ResolverModeCompare ResolverMode = "compare"
)

// IsValid reports whether m is a known resolver mode.
func (m ResolverMode) IsValid() bool {
	switch m {
	case ResolverModeFirstSuccess, ResolverModeAll, ResolverModeQuorum, ResolverModeCompare:
		return true
	}
	return false
}

// String returns the mode value for display/serialization.
func (m ResolverMode) String() string {
	return string(m)
}

// protocolFromResolver infers the transport protocol from a resolver string
// such as "udp:1.1.1.1:53", "tcp:8.8.8.8:53", "doh:https://..." or "dot:...".
func protocolFromResolver(resolver string) string {
	if strings.HasPrefix(resolver, "doh:") {
		return "doh"
	}
	if strings.HasPrefix(resolver, "dot:") {
		return "dot"
	}
	if strings.HasPrefix(resolver, "tcp:") {
		return "tcp"
	}
	return "udp"
}

type MarshalOption func(d *ResponseData)

func WithoutAllRecords() MarshalOption {
	return func(d *ResponseData) {
		d.AllRecords = nil
	}
}

func (d *ResponseData) JSON(options ...MarshalOption) (string, error) {
	dataToMarshal := *d
	// Always remove RawResp from JSON output as it's redundant and bloats the output
	dataToMarshal.RawResp = nil
	for _, option := range options {
		option(&dataToMarshal)
	}
	b, err := json.Marshal(dataToMarshal)
	return string(b), err
}

// DefaultOptions contains the default configuration options
var DefaultOptions = Options{
	BaseResolvers:     DefaultResolvers,
	MaxRetries:        5,
	QuestionTypes:     []uint16{miekgdns.TypeA},
	TraceMaxRecursion: 255,
	Hostsfile:         true,
	Timeout:           3 * time.Second,
}

// DefaultResolvers contains the list of resolvers known to be trusted.
var DefaultResolvers = []string{
	"udp:1.1.1.1:53",         // Cloudflare
	"udp:1.0.0.1:53",         // Cloudflare
	"udp:8.8.8.8:53",         // Google
	"udp:8.8.4.4:53",         // Google
	"udp:9.9.9.9:53",         // Quad9
	"udp:149.112.112.112:53", // Quad9
	"udp:208.67.222.222:53",  // Open DNS
	"udp:208.67.220.220:53",  // Open DNS
}

// New creates a dns resolver
func New(options Options) (*DNSX, error) {
	retryablednsOptions := retryabledns.Options{
		BaseResolvers: options.BaseResolvers,
		MaxRetries:    options.MaxRetries,
		Hostsfile:     options.Hostsfile,
		Proxy:         options.Proxy,
		Timeout:       options.Timeout,
	}

	dnsClient, err := retryabledns.NewWithOptions(retryablednsOptions)
	if err != nil {
		return nil, err
	}
	dnsClient.TCPFallback = true
	dnsx := &DNSX{dnsClient: dnsClient, Options: &options}
	if options.OutputCDN {
		dnsx.cdn = cdncheck.New()
	}
	if options.ResolverRateLimit > 0 {
		dnsx.resolverLimiters = make(map[string]*ratelimit.Limiter, len(options.BaseResolvers))
		for _, r := range options.BaseResolvers {
			if r == "" {
				continue
			}
			dnsx.resolverLimiters[r] = ratelimit.New(context.Background(), uint(options.ResolverRateLimit), time.Second)
		}
	}
	return dnsx, nil
}

// Lookup performs a DNS A question and returns corresponding IPs
func (d *DNSX) Lookup(hostname string) ([]string, error) {
	if iputil.IsIP(hostname) {
		return []string{hostname}, nil
	}

	dnsdata, err := d.dnsClient.Resolve(hostname)
	if err != nil {
		return nil, err
	}

	if dnsdata == nil || len(dnsdata.A) == 0 {
		return []string{}, errors.New("no ips found")
	}

	return dnsdata.A, nil
}

// QueryOne performs a DNS question of a specified type and returns raw responses
func (d *DNSX) QueryOne(hostname string) (*retryabledns.DNSData, error) {
	return d.dnsClient.Query(hostname, d.Options.QuestionTypes[0])
}

// QueryMultiple performs a DNS question of the specified types and returns raw responses
func (d *DNSX) QueryMultiple(hostname string) (*retryabledns.DNSData, error) {
	// Omit PTR queries unless the input is an IP address to decrease execution time, as PTR queries can lead to timeouts.
	filteredQuestionTypes := d.Options.QuestionTypes
	if d.Options.QueryAll {
		isIP := iputil.IsIP(hostname)
		if !isIP {
			filteredQuestionTypes = sliceutil.PruneEqual(filteredQuestionTypes, miekgdns.TypePTR)
		} else {
			filteredQuestionTypes = []uint16{miekgdns.TypePTR}
		}
	}
	return d.dnsClient.QueryMultiple(hostname, filteredQuestionTypes)
}

// QueryMultipleAll queries every configured resolver individually and returns
// one ResponseData per resolver, each carrying provenance (resolver, protocol,
// latency, attempts, RCODE). Implements the "all" resolver mode (issue #884).
func (d *DNSX) QueryMultipleAll(hostname string) ([]*ResponseData, error) {
	// Omit PTR queries unless the input is an IP address.
	filteredQuestionTypes := d.Options.QuestionTypes
	if d.Options.QueryAll {
		isIP := iputil.IsIP(hostname)
		if !isIP {
			filteredQuestionTypes = sliceutil.PruneEqual(filteredQuestionTypes, miekgdns.TypePTR)
		} else {
			filteredQuestionTypes = []uint16{miekgdns.TypePTR}
		}
	}

	resolvers := d.Options.BaseResolvers
	if len(resolvers) == 0 {
		resolvers = DefaultResolvers
	}

	var results []*ResponseData
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, rawResolver := range resolvers {
		wg.Add(1)
		go func(resolverStr string) {
			defer wg.Done()
			resolver := retryabledns.NetworkResolver{}
			// Parse resolver string: "udp:host:port", "host:port" or "host".
			// Defaults to UDP and port 53 when omitted.
			switch parts := strings.SplitN(resolverStr, ":", 3); len(parts) {
			case 3:
				resolver.Protocol = retryabledns.Protocol(parts[0])
				resolver.Host = parts[1]
				resolver.Port = parts[2]
			case 2:
				resolver.Protocol = retryabledns.UDP
				resolver.Host = parts[0]
				resolver.Port = parts[1]
			default:
				resolver.Protocol = retryabledns.UDP
				resolver.Host = resolverStr
				resolver.Port = "53"
			}

			queryStart := time.Now()
			// Per-resolver rate limit before contacting the resolver.
			if d.Options.ResolverRateLimit > 0 && d.resolverLimiters != nil {
				if lim, ok := d.resolverLimiters[resolverStr]; ok {
					lim.Take()
				}
			}
			dnsData, err := d.dnsClient.QueryMultipleWithResolver(hostname, filteredQuestionTypes, resolver)
			latency := time.Since(queryStart).Round(time.Millisecond).String()

			if err != nil || dnsData == nil {
				// Record a failed resolver probe so provenance is not lost.
				mu.Lock()
				results = append(results, &ResponseData{
					DNSData:  &retryabledns.DNSData{Host: hostname, Timestamp: time.Now()},
					Resolver: resolverStr,
					Protocol: protocolFromResolver(resolverStr),
					Latency:  latency,
					Attempts: 1,
				})
				mu.Unlock()
				return
			}

			// Ensure provenance fields are populated even when retryabledns
			// leaves them empty (e.g. empty answer from a resolver).
			if dnsData.Host == "" {
				dnsData.Host = hostname
			}
			if dnsData.Timestamp.IsZero() {
				dnsData.Timestamp = time.Now()
			}

			rd := &ResponseData{
				DNSData:   dnsData,
				Resolver:  resolverStr,
				Protocol:  protocolFromResolver(resolverStr),
				Latency:   latency,
				Attempts:  1,
				QueryTime: latency,
			}
			// If the resolver string carried protocol info, DNSData.Resolver may
			// already be populated; prefer the canonical resolver string.
			if len(dnsData.Resolver) == 0 {
				dnsData.Resolver = []string{resolverStr}
			}

			mu.Lock()
			results = append(results, rd)
			mu.Unlock()
		}(rawResolver)
	}

	wg.Wait()
	return results, nil
}

// contentSignature normalizes a DNS answer into a deterministic signature for
// quorum/compare grouping. It ignores TTL (per DNSSEC guidance: different TTLs
// alone are not content divergence) and record ordering, but considers names,
// types and values. Returns the signature string and whether the answer is
// "empty" (NOERROR with no records).
func contentSignature(data *retryabledns.DNSData) (string, bool) {
	// Build a stable representation of all non-empty record sets.
	type rrSet struct {
		typ    string
		values []string
	}
	var sets []rrSet

	addSet := func(typ string, vals []string) {
		if len(vals) == 0 {
			return
		}
		sorted := append([]string(nil), vals...)
		sort.Strings(sorted)
		sets = append(sets, rrSet{typ: typ, values: sorted})
	}

	addSet("A", data.A)
	addSet("AAAA", data.AAAA)
	addSet("CNAME", data.CNAME)
	addSet("MX", data.MX)
	addSet("NS", data.NS)
	addSet("PTR", data.PTR)
	addSet("SOA", soaStrings(data.SOA))
	addSet("TXT", data.TXT)
	addSet("SRV", data.SRV)
	addSet("CAA", data.CAA)

	sort.Slice(sets, func(i, j int) bool { return sets[i].typ < sets[j].typ })

	var b strings.Builder
	b.WriteString(fmt.Sprintf("rcode=%d", data.StatusCodeRaw))
	for _, s := range sets {
		b.WriteByte('|')
		b.WriteString(s.typ)
		b.WriteByte(':')
		b.WriteString(strings.Join(s.values, ","))
	}
	empty := len(sets) == 0
	return b.String(), empty
}

// soaStrings converts SOA records to a comparable string slice.
func soaStrings(soas []retryabledns.SOA) []string {
	out := make([]string, 0, len(soas))
	for _, s := range soas {
		out = append(out, fmt.Sprintf("%s|%s|%s|%d|%d|%d|%d|%d",
			s.Name, s.NS, s.Mbox, s.Serial, s.Refresh, s.Retry, s.Expire, s.Minttl))
	}
	return out
}

// QuorumResult is the outcome of a quorum query.
type QuorumResult struct {
	// ConsensusSignature is the content signature that reached quorum, or empty
	// if no signature reached the required votes.
	ConsensusSignature string `json:"consensus_signature,omitempty"`
	// ConsensusAnswer is the representative answer for the consensus signature,
	// or nil if no consensus was reached.
	ConsensusAnswer *ResponseData `json:"consensus_answer,omitempty"`
	// VoteCounts maps each content signature to the number of resolvers that
	// produced it.
	VoteCounts map[string]int `json:"vote_counts,omitempty"`
	// TotalVoters is the number of resolvers that produced any answer (including
	// errors/empty).
	TotalVoters int `json:"total_voters,omitempty"`
	// QuorumRequired is the minimum votes needed for consensus.
	QuorumRequired int `json:"quorum_required,omitempty"`
	// ConsensusReached is true when some signature reached QuorumRequired.
	ConsensusReached bool `json:"consensus_reached,omitempty"`
	// AllAnswers is every per-resolver answer with provenance.
	AllAnswers []*ResponseData `json:"all_answers,omitempty"`
}

// QueryMultipleQuorum queries every configured resolver and classifies answers
// by content consensus. Implements the "quorum" resolver mode. A signature
// reaches consensus when its vote count >= quorum. TTL differences alone do not
// count as content divergence.
func (d *DNSX) QueryMultipleQuorum(hostname string, quorum int) (*QuorumResult, error) {
	allAnswers, err := d.QueryMultipleAll(hostname)
	if err != nil {
		return nil, err
	}
	if quorum < 1 {
		quorum = 1
	}

	result := &QuorumResult{
		VoteCounts:     make(map[string]int),
		QuorumRequired: quorum,
		AllAnswers:     allAnswers,
	}

	for _, ans := range allAnswers {
		if ans.DNSData == nil {
			result.VoteCounts["<error>"]++
			result.TotalVoters++
			continue
		}
		sig, _ := contentSignature(ans.DNSData)
		result.VoteCounts[sig]++
		result.TotalVoters++
	}

	// Find the signature with the most votes that also meets quorum.
	bestSig := ""
	bestVotes := 0
	for sig, votes := range result.VoteCounts {
		if votes > bestVotes || (votes == bestVotes && sig < bestSig) {
			bestSig = sig
			bestVotes = votes
		}
	}

	if bestVotes >= quorum && bestSig != "<error>" {
		result.ConsensusReached = true
		result.ConsensusSignature = bestSig
		// Pick the first answer matching the best signature as representative.
		for _, ans := range allAnswers {
			if ans.DNSData != nil {
				sig, _ := contentSignature(ans.DNSData)
				if sig == bestSig {
					result.ConsensusAnswer = ans
					break
				}
			}
		}
	}

	return result, nil
}

// CompareResult is the outcome of a compare query.
type CompareResult struct {
	// AllAnswers is every per-resolver answer with provenance, in resolver
	// order (not arrival order).
	AllAnswers []*ResponseData `json:"all_answers,omitempty"`
	// Divergences lists content signatures produced by more than one resolver
	// but with differing values, plus signatures produced by only one resolver.
	Divergences []CompareDivergence `json:"divergences,omitempty"`
	// TotalResolvers is the number of resolvers queried.
	TotalResolvers int `json:"total_resolvers,omitempty"`
	// SuccessfulResolvers is the number that returned a parseable answer.
	SuccessfulResolvers int `json:"successful_resolvers,omitempty"`
	// FailedResolvers is the number that errored or returned empty.
	FailedResolvers int `json:"failed_resolvers,omitempty"`
}

// CompareDivergence describes one content-signature divergence between resolvers.
type CompareDivergence struct {
	// Signature is the normalized content signature.
	Signature string `json:"signature,omitempty"`
	// Resolvers lists the resolvers that produced this signature.
	Resolvers []string `json:"resolvers,omitempty"`
	// VoteCount is how many resolvers produced exactly this signature.
	VoteCount int `json:"vote_count,omitempty"`
	// Unique indicates the signature was produced by only one resolver.
	Unique bool `json:"unique,omitempty"`
}

// QueryMultipleCompare queries every configured resolver and returns every
// answer with provenance, plus a structured list of divergences. Implements the
// "compare" resolver mode. Comparison is independent of response arrival order.
func (d *DNSX) QueryMultipleCompare(hostname string) (*CompareResult, error) {
	allAnswers, err := d.QueryMultipleAll(hostname)
	if err != nil {
		return nil, err
	}

	result := &CompareResult{
		AllAnswers:     allAnswers,
		TotalResolvers: len(allAnswers),
		Divergences:    []CompareDivergence{},
	}

	sigToResolvers := make(map[string][]string)
	for _, ans := range allAnswers {
		if ans.DNSData == nil {
			result.FailedResolvers++
			continue
		}
		result.SuccessfulResolvers++
		sig, _ := contentSignature(ans.DNSData)
		sigToResolvers[sig] = append(sigToResolvers[sig], ans.Resolver)
	}

	for sig, resolvers := range sigToResolvers {
		div := CompareDivergence{
			Signature: sig,
			Resolvers: resolvers,
			VoteCount: len(resolvers),
			Unique:    len(resolvers) == 1,
		}
		result.Divergences = append(result.Divergences, div)
	}

	// Sort divergences for deterministic output: non-unique first, then by
	// signature string.
	sort.Slice(result.Divergences, func(i, j int) bool {
		if result.Divergences[i].Unique != result.Divergences[j].Unique {
			return !result.Divergences[i].Unique // non-unique (divergent) first
		}
		return result.Divergences[i].Signature < result.Divergences[j].Signature
	})

	return result, nil
}

// Trace performs a DNS trace of the specified types and returns raw responses
func (d *DNSX) Trace(hostname string) (*retryabledns.TraceData, error) {
	if len(d.Options.QuestionTypes) == 0 {
		return nil, errors.New("no question types specified for trace")
	}
	return d.dnsClient.Trace(hostname, d.Options.QuestionTypes[0], d.Options.TraceMaxRecursion)
}

// Trace performs a DNS trace of the specified types and returns raw responses
func (d *DNSX) AXFR(hostname string) (*retryabledns.AXFRData, error) {
	return d.dnsClient.AXFR(hostname)
}
