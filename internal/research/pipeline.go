package research

import (
	"net"
	"sync"

	"github.com/projectdiscovery/dnsx/libs/dnsx"
)

// PipelineConfig controls the fused resolve -> probe flow.
type PipelineConfig struct {
	// Threads is the number of concurrent target workers.
	Threads int
	// DefaultSchemes are probed when a target carries no explicit scheme.
	DefaultSchemes []string
	// Resolve enables the DNS stage. When false, targets are probed directly
	// (useful when inputs are already URLs/IPs).
	Resolve bool
	// CDNCheck enriches resolved hosts with CDN/WAF classification, reusing the
	// httpx engine's cdncheck. Requires Resolve.
	CDNCheck bool
}

// Pipeline fuses the DNS resolution engine with the HTTP probe engine. A
// resolver of nil is valid only when cfg.Resolve is false.
type Pipeline struct {
	resolver *dnsx.DNSX
	prober   *Prober
	cfg      PipelineConfig
}

// NewPipeline wires a resolver and prober into a pipeline. DefaultSchemes
// falls back to https,http and Threads to 1 when unset.
func NewPipeline(resolver *dnsx.DNSX, prober *Prober, cfg PipelineConfig) *Pipeline {
	if cfg.Threads <= 0 {
		cfg.Threads = 1
	}
	if len(cfg.DefaultSchemes) == 0 {
		cfg.DefaultSchemes = []string{"https", "http"}
	}
	return &Pipeline{resolver: resolver, prober: prober, cfg: cfg}
}

// Run processes every input concurrently and invokes emit for each produced
// Result. emit is serialized: the pipeline guarantees it is never called from
// two goroutines at once, so callers need no locking of their own.
func (p *Pipeline) Run(inputs []string, emit func(Result)) {
	var emitMu sync.Mutex
	safeEmit := func(r Result) {
		emitMu.Lock()
		emit(r)
		emitMu.Unlock()
	}

	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < p.cfg.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for raw := range jobs {
				p.process(raw, safeEmit)
			}
		}()
	}
	for _, in := range inputs {
		jobs <- in
	}
	close(jobs)
	wg.Wait()
}

// process handles a single raw input: parse -> (resolve) -> probe -> emit.
func (p *Pipeline) process(raw string, emit func(Result)) {
	target, err := parseTarget(raw)
	if err != nil {
		emit(Result{Input: raw, Failed: true, Error: err.Error()})
		return
	}

	base := Result{Input: raw, Host: target.Host}
	if p.cfg.Resolve && p.resolver != nil && net.ParseIP(target.Host) == nil {
		base = p.resolve(base, target.Host)
		if len(base.A) == 0 && len(base.AAAA) == 0 {
			// No address records: record the DNS outcome, skip HTTP.
			base.Failed = true
			if base.Error == "" {
				base.Error = "no DNS address records"
			}
			emit(base)
			return
		}
		if p.cfg.CDNCheck && len(base.A) > 0 {
			base.IsCDN, base.CDNName, base.CDNType = p.prober.CheckCDN(base.A[0])
		}
	}

	for _, pu := range target.probeURLs(p.cfg.DefaultSchemes) {
		emit(p.prober.Probe(base, pu.url, pu.scheme, pu.port))
	}
}

// resolve runs the DNS stage and folds the answers into base. It uses
// QueryMultiple (first-success across configured resolvers), which is the
// project's default dispatch mode and returns the combined A/AAAA/CNAME set.
func (p *Pipeline) resolve(base Result, host string) Result {
	data, err := p.resolver.QueryMultiple(host)
	if err != nil {
		base.Error = err.Error()
		return base
	}
	if data == nil {
		base.Error = "nil DNS response"
		return base
	}
	base.A = appendUnique(base.A, data.A...)
	base.AAAA = appendUnique(base.AAAA, data.AAAA...)
	base.CNAME = appendUnique(base.CNAME, data.CNAME...)
	if len(data.Resolver) > 0 {
		base.DNSResolver = data.Resolver[0]
	}
	return base
}

// appendUnique appends values not already present, preserving order.
func appendUnique(dst []string, values ...string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, v := range dst {
		seen[v] = struct{}{}
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		dst = append(dst, v)
	}
	return dst
}
