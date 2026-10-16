package main

import (
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
)

// localServer answers A queries for "host{i}.example.test" with 127.0.0.1.
type localServer struct {
	packetConn net.PacketConn
	addr       string
}

func newLocalServer() (*localServer, error) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	addr := pc.LocalAddr().String()
	s := &localServer{packetConn: pc, addr: addr}
	go s.serve()
	return s, nil
}

func (s *localServer) serve() {
	buf := make([]byte, 512)
	for {
		n, clientAddr, err := s.packetConn.ReadFrom(buf)
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
		q := reply.Question[0]
		if q.Qtype == dns.TypeA {
			reply.Answer = append(reply.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP("127.0.0.1"),
			})
		}
		out, _ := reply.Pack()
		s.packetConn.WriteTo(out, clientAddr)
	}
}

func (s *localServer) close() {
	s.packetConn.Close()
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(p / 100.0 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

func runMode(dx *dnsx.DNSX, mode dnsx.ResolverMode, hosts []string) (successful int, latencies []time.Duration, total time.Duration) {
	latencies = make([]time.Duration, 0, len(hosts))
	results := make([]*dnsx.ResponseData, len(hosts))
	var wg sync.WaitGroup
	start := time.Now()
	for i, h := range hosts {
		wg.Add(1)
		go func(idx int, host string) {
			defer wg.Done()
			var ans []*dnsx.ResponseData
			var err error
			queryStart := time.Now()
			switch mode {
			case dnsx.ResolverModeFirstSuccess:
				single, e := dx.QueryMultiple(host)
				if e == nil && single != nil {
					ans = []*dnsx.ResponseData{{DNSData: single}}
				}
				err = e
			case dnsx.ResolverModeAll:
				ans, err = dx.QueryMultipleAll(host)
			case dnsx.ResolverModeQuorum:
				qr, e := dx.QueryMultipleQuorum(host, 1)
				if e == nil && qr != nil && qr.ConsensusAnswer != nil {
					ans = []*dnsx.ResponseData{qr.ConsensusAnswer}
				}
				err = e
			case dnsx.ResolverModeCompare:
				cr, e := dx.QueryMultipleCompare(host)
				if e == nil && cr != nil {
					ans = cr.AllAnswers
				}
				err = e
			}
			if err != nil || len(ans) == 0 {
				results[idx] = nil
				return
			}
			results[idx] = ans[0]
			elapsed := time.Since(queryStart)
			if d, perr := time.ParseDuration(ans[0].Latency); perr != nil || d == 0 {
				latencies = append(latencies, elapsed)
			} else {
				latencies = append(latencies, d)
			}
		}(i, h)
	}
	wg.Wait()
	total = time.Since(start)
	for _, r := range results {
		if r != nil {
			successful++
		}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	return successful, latencies, total
}

func main() {
	server, err := newLocalServer()
	if err != nil {
		panic(err)
	}
	defer server.close()
	time.Sleep(50 * time.Millisecond)

	hostCount := 1000
	hosts := make([]string, hostCount)
	for i := 0; i < hostCount; i++ {
		hosts[i] = fmt.Sprintf("host%d.example.test", i)
	}

	modes := []struct {
		name string
		mode dnsx.ResolverMode
	}{
		{"first-success", dnsx.ResolverModeFirstSuccess},
		{"all", dnsx.ResolverModeAll},
		{"quorum(1)", dnsx.ResolverModeQuorum},
		{"compare", dnsx.ResolverModeCompare},
	}

	fmt.Println("Benchmark: local DNS server (127.0.0.1), 1 resolver, 1000 hosts, type A")
	fmt.Println("================================================================================")
	fmt.Printf("%-18s %12s %12s %12s %12s\n", "mode", "total(s)", "qps", "successful", "p50(ms)")

	for _, m := range modes {
		dx, err := dnsx.New(dnsx.Options{
			BaseResolvers: []string{server.addr},
			QuestionTypes: []uint16{dns.TypeA},
			Timeout:       2 * time.Second,
			MaxRetries:    1,
			ResolverMode:  m.mode,
		})
		if err != nil {
			panic(err)
		}
		successful, latencies, total := runMode(dx, m.mode, hosts)
		qps := float64(successful) / total.Seconds()
		p50 := percentile(latencies, 0.50)
		fmt.Printf("%-18s %12.4f %12.1f %12d %12.1f\n",
			m.name, total.Seconds(), qps, successful, float64(p50)/float64(time.Millisecond))
	}

	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("- qps = successful answers / total wall-clock time.")
	fmt.Println("- p50 computed over per-host latency; falls back to per-call wall time when")
	fmt.Println("  the resolver mode does not populate Latency (first-success).")
	fmt.Println("- No public targets used; all traffic stays on 127.0.0.1.")
	fmt.Println("- first-success is the upstream-compatible baseline.")
	fmt.Println("- 'all', 'quorum', 'compare' are HyperDNSx additions with per-resolver provenance.")
}
