// Command hyperx is the HyperDNSx deep-research pipeline: it fuses the DNS
// resolution engine with the httpx HTTP probing engine so a single pass takes
// raw targets to resolved, fingerprinted HTTP endpoints emitted as unified
// JSONL. The pure-DNS `dnsx` binary is left untouched.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/projectdiscovery/dnsx/internal/research"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

// resolveQuestionTypes are the DNS record types the research resolver asks for:
// A/AAAA give addresses to probe, CNAME preserves the alias chain for output.
var resolveQuestionTypes = []uint16{dns.TypeA, dns.TypeAAAA, dns.TypeCNAME}

// defaultResolvers are used when the caller supplies none.
var defaultResolvers = []string{
	"1.1.1.1:53",
	"8.8.8.8:53",
	"8.8.4.4:53",
	"1.0.0.1:53",
}

type options struct {
	list            string
	output          string
	resolvers       string
	schemes         string
	threads         int
	timeout         int
	retries         int
	followRedirects bool
	tlsGrab         bool
	tlsImpersonate  string
	randomAgent     bool
	noResolve       bool
	extractJS       bool
	extractEmails   bool
	extractRegex    string
	jsonOut         bool
	silent          bool
}

func main() {
	opts := parseOptions()

	if opts.silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	} else {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelInfo)
	}

	inputs, err := readInputs(opts.list)
	if err != nil {
		gologger.Fatal().Msgf("could not read inputs: %s", err)
	}
	if len(inputs) == 0 {
		gologger.Fatal().Msgf("no targets provided (use -l or pipe via stdin)")
	}

	regexes, err := research.CompileRegexes(splitComma(opts.extractRegex))
	if err != nil {
		gologger.Fatal().Msgf("invalid -extract-regex: %s", err)
	}
	extractor := research.NewExtractor(research.ExtractConfig{
		JS:      opts.extractJS,
		Emails:  opts.extractEmails,
		Regexes: regexes,
	})

	prober, err := research.NewProber(research.ProbeConfig{
		Timeout:         time.Duration(opts.timeout) * time.Second,
		RetryMax:        opts.retries,
		FollowRedirects: opts.followRedirects,
		TLSGrab:         opts.tlsGrab,
		TLSImpersonate:  opts.tlsImpersonate,
		RandomAgent:     opts.randomAgent,
		Threads:         opts.threads,
		Extractor:       extractor,
	})
	if err != nil {
		gologger.Fatal().Msgf("could not init probe engine: %s", err)
	}

	var resolver *dnsx.DNSX
	if !opts.noResolve {
		resolver, err = dnsx.New(dnsx.Options{
			BaseResolvers: resolverList(opts.resolvers),
			MaxRetries:    opts.retries,
			Timeout:       time.Duration(opts.timeout) * time.Second,
			QuestionTypes: resolveQuestionTypes,
		})
		if err != nil {
			gologger.Fatal().Msgf("could not init resolver: %s", err)
		}
	}

	pipeline := research.NewPipeline(resolver, prober, research.PipelineConfig{
		Threads:        opts.threads,
		DefaultSchemes: schemeList(opts.schemes),
		Resolve:        !opts.noResolve,
	})

	writer, closeWriter, err := outputWriter(opts.output)
	if err != nil {
		gologger.Fatal().Msgf("could not open output: %s", err)
	}
	defer closeWriter()

	var found int
	pipeline.Run(inputs, func(r research.Result) {
		line := formatResult(r, opts.jsonOut)
		if line == "" {
			return
		}
		found++
		fmt.Fprintln(writer, line)
	})
	gologger.Info().Msgf("deep-research complete: %d records over %d targets", found, len(inputs))
}

func parseOptions() *options {
	o := &options{}
	fs := goflags.NewFlagSet()
	fs.SetDescription("hyperx fuses DNS resolution with HTTP probing into a single deep-research pass.")

	fs.CreateGroup("input", "Input",
		fs.StringVarP(&o.list, "list", "l", "", "targets to research (file, comma-separated, or stdin)"),
	)
	fs.CreateGroup("probe", "Probe",
		fs.StringVar(&o.schemes, "schemes", "https,http", "schemes to probe when target has none"),
		fs.BoolVar(&o.followRedirects, "follow-redirects", false, "follow HTTP redirects"),
		fs.BoolVar(&o.tlsGrab, "tls-grab", false, "grab TLS certificate + JA3/JARM data"),
		fs.StringVar(&o.tlsImpersonate, "tls-impersonate", "", "ClientHello profile to impersonate (e.g. chrome)"),
		fs.BoolVar(&o.randomAgent, "random-agent", false, "use a random User-Agent per request"),
		fs.BoolVar(&o.noResolve, "no-resolve", false, "skip the DNS stage and probe targets directly"),
	)
	fs.CreateGroup("extract", "Extract",
		fs.BoolVar(&o.extractJS, "extract-js", false, "extract referenced .js files from responses"),
		fs.BoolVar(&o.extractEmails, "extract-emails", false, "extract email addresses from responses"),
		fs.StringVar(&o.extractRegex, "extract-regex", "", "custom extraction regexes (comma-separated name=pattern)"),
	)
	fs.CreateGroup("resolver", "Resolver",
		fs.StringVarP(&o.resolvers, "resolvers", "r", "", "DNS resolvers (file or comma-separated; default public resolvers)"),
	)
	fs.CreateGroup("rate", "Rate",
		fs.IntVarP(&o.threads, "threads", "t", 25, "number of concurrent workers"),
		fs.IntVar(&o.timeout, "timeout", 10, "per-request timeout in seconds"),
		fs.IntVar(&o.retries, "retries", 2, "number of retries"),
	)
	fs.CreateGroup("output", "Output",
		fs.StringVarP(&o.output, "output", "o", "", "file to write results to"),
		fs.BoolVarP(&o.jsonOut, "json", "j", false, "emit results as JSONL"),
		fs.BoolVar(&o.silent, "silent", false, "only print results, no status logs"),
	)

	if err := fs.Parse(); err != nil {
		gologger.Fatal().Msgf("could not parse flags: %s", err)
	}
	return o
}

// readInputs collects targets from a file, a comma-separated string, or stdin.
func readInputs(list string) ([]string, error) {
	list = strings.TrimSpace(list)
	if list != "" {
		if info, err := os.Stat(list); err == nil && !info.IsDir() {
			return readLines(list)
		}
		var out []string
		for _, p := range strings.Split(list, ",") {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out, nil
	}
	return readStdin()
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()
	return scanLines(f)
}

func readStdin() ([]string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat stdin: %w", err)
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, nil // interactive terminal, no piped input
	}
	return scanLines(os.Stdin)
}

func scanLines(f *os.File) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan input: %w", err)
	}
	return out, nil
}

// outputWriter returns a writer for stdout or the given file, plus a closer.
func outputWriter(path string) (*bufio.Writer, func(), error) {
	if strings.TrimSpace(path) == "" {
		w := bufio.NewWriter(os.Stdout)
		return w, func() { w.Flush() }, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("create %q: %w", path, err)
	}
	w := bufio.NewWriter(f)
	return w, func() { w.Flush(); f.Close() }, nil
}

// formatResult renders a Result as JSON or a human-readable line. Failed probes
// are suppressed in text mode but kept in JSON mode for completeness.
func formatResult(r research.Result, asJSON bool) string {
	if asJSON {
		line, err := r.JSON()
		if err != nil {
			gologger.Warning().Msgf("marshal result for %q: %s", r.Input, err)
			return ""
		}
		return line
	}
	if r.Failed || r.StatusCode == 0 {
		return ""
	}
	parts := []string{r.URL}
	parts = append(parts, fmt.Sprintf("[%d]", r.StatusCode))
	if r.Title != "" {
		parts = append(parts, fmt.Sprintf("[%s]", r.Title))
	}
	if r.WebServer != "" {
		parts = append(parts, fmt.Sprintf("[%s]", r.WebServer))
	}
	return strings.Join(parts, " ")
}

func resolverList(s string) []string {
	parsed := splitComma(s)
	if len(parsed) == 0 {
		return defaultResolvers
	}
	return parsed
}

func schemeList(s string) []string {
	parsed := splitComma(s)
	if len(parsed) == 0 {
		return []string{"https", "http"}
	}
	return parsed
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
