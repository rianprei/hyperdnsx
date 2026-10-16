package dnsx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	miekgdns "github.com/miekg/dns"
	"github.com/projectdiscovery/cdncheck"
	retryabledns "github.com/projectdiscovery/retryabledns"
	iputil "github.com/projectdiscovery/utils/ip"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

// DNSX is structure to perform dns lookups
type DNSX struct {
	dnsClient *retryabledns.Client
	Options   *Options
	cdn       *cdncheck.Client
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
