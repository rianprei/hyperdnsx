package runner

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/logrusorgru/aurora/v4"
	"github.com/miekg/dns"
	"github.com/projectdiscovery/dnsx/libs/dnsx"
	"github.com/projectdiscovery/goconfig"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/formatter"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/utils/auth/pdcp"
	"github.com/projectdiscovery/utils/env"
	fileutil "github.com/projectdiscovery/utils/file"
	updateutils "github.com/projectdiscovery/utils/update"
)

const (
	DefaultResumeFile = "resume.cfg"
)

var PDCPApiKey string

type Options struct {
	Resolvers             string
	Hosts                 string
	Domains               string
	WordList              string
	Threads               int
	RateLimit             int
	Retries               int
	OutputFormat          string
	OutputTemplate        string
	OutputFile            string
	Raw                   bool
	Silent                bool
	Verbose               bool
	Version               bool
	NoColor               bool
	Response              bool
	ResponseOnly          bool
	A                     bool
	AAAA                  bool
	NS                    bool
	CNAME                 bool
	PTR                   bool
	MX                    bool
	SOA                   bool
	ANY                   bool
	TXT                   bool
	SRV                   bool
	AXFR                  bool
	JSON                  bool
	OmitRaw               bool
	Trace                 bool
	TraceMaxRecursion     int
	WildcardThreshold     int
	AutoWildcard          bool
	WildcardDomain        string
	ShowStatistics        bool
	rcodes                map[int]struct{}
	RCode                 string
	ResponseTypeFilter    string
	responseTypeFilterMap []string
	explicitRecordTypes   bool
	hasRCodes             bool
	Resume                bool
	resumeCfg             *ResumeCfg
	HostsFile             bool
	Stream                bool
	Timeout               time.Duration
	CAA                   bool
	QueryAll              bool
	ExcludeType           []string
	OutputCDN             bool
	ASN                   bool
	HealthCheck           bool
	DisableUpdateCheck    bool
	PdcpAuth              string
	Proxy                 string
	ResolverMode          string
	resolverMode          dnsx.ResolverMode
	Quorum                int
	quorum                int
	DryRun                bool
	ScopeFile             string
	scope                 []string
	scopeLoaded           bool
	ResolverRateLimit     int
}

// ShouldLoadResume resume file
func (options *Options) ShouldLoadResume() bool {
	return options.Resume && fileutil.FileExists(DefaultResumeFile)
}

// ShouldSaveResume file
func (options *Options) ShouldSaveResume() bool {
	return true
}

// ParseOptions parses the command line options for application
func ParseOptions() *Options {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`dnsx is a fast and multi-purpose DNS toolkit allow to run multiple probes using retryabledns library.`)

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.Hosts, "list", "l", "", "list of sub(domains)/hosts to resolve (file or comma separated or stdin)"),
		flagSet.StringVarP(&options.Domains, "domain", "d", "", "list of domain to bruteforce (file or comma separated or stdin)"),
		flagSet.StringVarP(&options.WordList, "wordlist", "w", "", "list of words to bruteforce (file or comma separated or stdin)"),
	)

	queries := goflags.AllowdTypes{
		"none":  goflags.EnumVariable(0),
		"a":     goflags.EnumVariable(1),
		"aaaa":  goflags.EnumVariable(2),
		"cname": goflags.EnumVariable(3),
		"ns":    goflags.EnumVariable(4),
		"txt":   goflags.EnumVariable(5),
		"srv":   goflags.EnumVariable(6),
		"ptr":   goflags.EnumVariable(7),
		"mx":    goflags.EnumVariable(8),
		"soa":   goflags.EnumVariable(9),
		"axfr":  goflags.EnumVariable(10),
		"caa":   goflags.EnumVariable(11),
		"any":   goflags.EnumVariable(12),
	}

	flagSet.CreateGroup("query", "Query",
		flagSet.BoolVar(&options.A, "a", false, "query A record (default)"),
		flagSet.BoolVar(&options.AAAA, "aaaa", false, "query AAAA record"),
		flagSet.BoolVar(&options.CNAME, "cname", false, "query CNAME record"),
		flagSet.BoolVar(&options.NS, "ns", false, "query NS record"),
		flagSet.BoolVar(&options.TXT, "txt", false, "query TXT record"),
		flagSet.BoolVar(&options.SRV, "srv", false, "query SRV record"),
		flagSet.BoolVar(&options.PTR, "ptr", false, "query PTR record"),
		flagSet.BoolVar(&options.MX, "mx", false, "query MX record"),
		flagSet.BoolVar(&options.SOA, "soa", false, "query SOA record"),
		flagSet.BoolVar(&options.ANY, "any", false, "query ANY record"),
		flagSet.BoolVar(&options.AXFR, "axfr", false, "query AXFR"),
		flagSet.BoolVar(&options.CAA, "caa", false, "query CAA record"),
		flagSet.BoolVarP(&options.QueryAll, "recon", "all", false, "query all the dns records (a,aaaa,cname,ns,txt,srv,ptr,mx,soa,axfr,caa)"),
		flagSet.EnumSliceVarP(&options.ExcludeType, "exclude-type", "e", []goflags.EnumVariable{0}, "dns query type to exclude (a,aaaa,cname,ns,txt,srv,ptr,mx,soa,axfr,caa)", queries),
	)

	flagSet.CreateGroup("filter", "Filter",
		flagSet.BoolVarP(&options.Response, "resp", "re", false, "display dns response"),
		flagSet.BoolVarP(&options.ResponseOnly, "resp-only", "ro", false, "display dns response only"),
		flagSet.StringVarP(&options.RCode, "rcode", "rc", "", "filter result by dns status code (eg. -rcode noerror,servfail,refused)"),
		flagSet.StringVarP(&options.ResponseTypeFilter, "response-type-filter", "rtf", "", "return entries with no records for the specified query types (e.g., a, cname)"),
	)

	flagSet.CreateGroup("probe", "Probe",
		flagSet.BoolVar(&options.OutputCDN, "cdn", false, "display cdn name"),
		flagSet.BoolVar(&options.ASN, "asn", false, "display host asn information"),
	)

	flagSet.CreateGroup("rate-limit", "Rate-limit",
		flagSet.IntVarP(&options.Threads, "threads", "t", 100, "number of concurrent threads to use"),
		flagSet.IntVarP(&options.RateLimit, "rate-limit", "rl", -1, "number of dns request/second to make (disabled as default)"),
		flagSet.IntVarP(&options.ResolverRateLimit, "resolver-rate-limit", "", -1, "per-resolver dns request/second limit (disabled as default)"),
	)

	flagSet.CreateGroup("update", "Update",
		flagSet.CallbackVarP(GetUpdateCallback(), "update", "up", "update dnsx to latest version"),
		flagSet.BoolVarP(&options.DisableUpdateCheck, "disable-update-check", "duc", false, "disable automatic dnsx update check"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.OutputFile, "output", "o", "", "file to write output"),
		flagSet.BoolVarP(&options.JSON, "json", "j", false, "write output in JSONL(ines) format"),
		flagSet.BoolVarP(&options.OmitRaw, "or", "omit-raw", false, "omit raw dns response from jsonl output"),
		flagSet.StringVarP(&options.OutputTemplate, "output-template", "ot", "", "custom output template (e.g. -ot '{{host}} {{a}}')"),
	)

	flagSet.CreateGroup("debug", "Debug",
		flagSet.BoolVarP(&options.HealthCheck, "health-check", "hc", false, "run diagnostic check up"),
		flagSet.BoolVar(&options.Silent, "silent", false, "display only results in the output"),
		flagSet.BoolVarP(&options.Verbose, "verbose", "v", false, "display verbose output"),
		flagSet.BoolVarP(&options.Raw, "debug", "raw", false, "display raw dns response"),
		flagSet.BoolVar(&options.ShowStatistics, "stats", false, "display stats of the running scan"),
		flagSet.BoolVar(&options.Version, "version", false, "display version of dnsx"),
		flagSet.BoolVarP(&options.NoColor, "no-color", "nc", false, "disable color in output"),
	)

	flagSet.CreateGroup("optimization", "Optimization",
		flagSet.IntVar(&options.Retries, "retry", 2, "number of dns attempts to make (must be at least 1)"),
		flagSet.BoolVarP(&options.HostsFile, "hostsfile", "hf", false, "use system host file"),
		flagSet.BoolVar(&options.Trace, "trace", false, "perform dns tracing"),
		flagSet.IntVar(&options.TraceMaxRecursion, "trace-max-recursion", 255, "Max recursion for dns trace"),
		flagSet.BoolVar(&options.Resume, "resume", false, "resume existing scan"),
		flagSet.BoolVar(&options.Stream, "stream", false, "stream mode (wordlist, wildcard, stats and stop/resume will be disabled)"),
		flagSet.DurationVar(&options.Timeout, "timeout", 3*time.Second, "maximum time to wait for a DNS query to complete"),
		flagSet.BoolVar(&options.DryRun, "dry-run", false, "show planned queries and estimates without sending any DNS traffic"),
	)

	flagSet.CreateGroup("configs", "Configurations",
		flagSet.DynamicVar(&options.PdcpAuth, "auth", "true", "configure ProjectDiscovery Cloud Platform (PDCP) api key"),
		flagSet.StringVarP(&options.Resolvers, "resolver", "r", "", "list of resolvers to use (file or comma separated)"),
		flagSet.StringVarP(&options.ResolverMode, "resolver-mode", "", "", "resolver dispatch mode: first-success (default), all, quorum, compare"),
		flagSet.IntVarP(&options.Quorum, "quorum", "", 0, "minimum votes for consensus in quorum mode (0 = majority of resolvers)"),
		flagSet.IntVarP(&options.WildcardThreshold, "wildcard-threshold", "wt", 5, "wildcard filter threshold"),
		flagSet.BoolVar(&options.AutoWildcard, "auto-wildcard", false, "automatically detect wildcard domains for filtering"),
		flagSet.StringVarP(&options.WildcardDomain, "wildcard-domain", "wd", "", "domain name for manual wildcard filtering (mutually exclusive with -auto-wildcard; other flags will be ignored - json output recommended)"),
		flagSet.StringVar(&options.Proxy, "proxy", "", "proxy to use (eg socks5://127.0.0.1:8080)"),
		flagSet.StringVarP(&options.ScopeFile, "scope-file", "", "", "file with allowed scope domains; candidates outside scope are rejected"),
	)

	_ = flagSet.Parse()

	if options.HealthCheck {
		gologger.Print().Msgf("%s\n", DoHealthCheck(options, flagSet))
		os.Exit(0)
	}

	normalizedWildcardDomain, err := normalizeAndValidateWildcardDomain(options.WildcardDomain)
	if err != nil {
		gologger.Fatal().Msgf("invalid wildcard-domain value")
	}
	options.WildcardDomain = normalizedWildcardDomain

	if options.ResponseTypeFilter != "" {
		filterTypes := strings.Split(options.ResponseTypeFilter, ",")
		// Clean and validate filter types
		validTypes := make([]string, 0, len(filterTypes))
		for _, et := range filterTypes {
			et = strings.TrimSpace(strings.ToLower(et))
			if et != "" {
				validTypes = append(validTypes, et)
			}
		}
		options.responseTypeFilterMap = validTypes
	}

	options.configureQueryOptions()

	err = options.configureRcodes()
	if err != nil {
		gologger.Fatal().Msgf("%s\n", err)
	}

	err = options.configureResume()
	if err != nil {
		gologger.Fatal().Msgf("%s\n", err)
	}

	// api key hierarchy: cli flag > env var > .pdcp/credential file
	if options.PdcpAuth == "true" {
		AuthWithPDCP()
	} else if len(options.PdcpAuth) == 36 {
		PDCPApiKey = options.PdcpAuth
		ph := pdcp.PDCPCredHandler{}
		if _, err := ph.GetCreds(); err == pdcp.ErrNoCreds {
			apiServer := env.GetEnvOrDefault("PDCP_API_SERVER", pdcp.DefaultApiServer)
			if validatedCreds, err := ph.ValidateAPIKey(PDCPApiKey, apiServer, "dnsx"); err == nil {
				_ = ph.SaveCreds(validatedCreds)
			}
		}
	}

	options.configureOutput()
	showBanner()

	options.configureResolverMode()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", version)
		os.Exit(0)
	}

	if !options.DisableUpdateCheck {
		latestVersion, err := updateutils.GetToolVersionCallback("dnsx", version)()
		if err != nil {
			if options.Verbose {
				gologger.Error().Msgf("dnsx version check failed: %v", err.Error())
			}
		} else {
			gologger.Info().Msgf("Current dnsx version %v %v", version, updateutils.GetVersionDescription(version, latestVersion))
		}
	}

	options.validateOptions()

	if !options.DryRun && options.ScopeFile != "" {
		if err := options.loadScope(); err != nil {
			gologger.Fatal().Msgf("scope-file: %v", err)
		}
	}

	if options.DryRun {
		RunDryRun(options)
		os.Exit(0)
	}

	return options
}

// RunDryRun prints the planned queries and estimates without sending any DNS
// traffic. It satisfies the --dry-run contract: no network connections, no
// update checks, no authentication, and no irreversible consumption of stdin.
func RunDryRun(options *Options) {
	gologger.Print().Msgf("=== HyperDNSx dry-run (no DNS traffic will be sent) ===")
	gologger.Print().Msgf("Resolver mode : %s", options.resolverMode)
	gologger.Print().Msgf("Resolvers     : %s", options.Resolvers)
	gologger.Print().Msgf("Timeout       : %s", options.Timeout)
	gologger.Print().Msgf("Retries       : %d", options.Retries)
	gologger.Print().Msgf("Rate limit    : %d req/s", options.RateLimit)
	gologger.Print().Msgf("Question types: %s", questionTypesSummary(options))
	gologger.Print().Msgf("Output JSON   : %v", options.JSON)

	// Estimate candidate count from inputs without consuming stdin
	// irreversibly: count file lines or comma-separated items.
	candidates := estimateCandidates(options)
	gologger.Print().Msgf("Candidates    : %d (estimated)", candidates)

	queryTypes := selectedQuestionTypes(options)
	if queryTypes == 0 {
		queryTypes = 1 // default A
	}
	resolverCount := len(strings.Split(options.Resolvers, ","))
	if resolverCount == 0 || options.Resolvers == "" {
		resolverCount = len(dnsx.DefaultResolvers)
	}
	multiplier := resolverCount
	if options.resolverMode == dnsx.ResolverModeFirstSuccess {
		multiplier = 1
	}
	estimatedQueries := candidates * queryTypes * multiplier
	gologger.Print().Msgf("Est. queries  : %d (candidates × types × resolvers/mode)", estimatedQueries)
	gologger.Print().Msgf("=== end dry-run ===")
}

func questionTypesSummary(options *Options) string {
	var types []string
	for _, t := range []struct {
		name string
		set  bool
	}{
		{"A", options.A}, {"AAAA", options.AAAA}, {"CNAME", options.CNAME},
		{"NS", options.NS}, {"TXT", options.TXT}, {"SRV", options.SRV},
		{"PTR", options.PTR}, {"MX", options.MX}, {"SOA", options.SOA},
		{"ANY", options.ANY}, {"AXFR", options.AXFR}, {"CAA", options.CAA},
	} {
		if t.set {
			types = append(types, t.name)
		}
	}
	if len(types) == 0 {
		return "A (default)"
	}
	return strings.Join(types, ", ")
}

func selectedQuestionTypes(options *Options) int {
	count := 0
	if options.A {
		count++
	}
	if options.AAAA {
		count++
	}
	if options.CNAME {
		count++
	}
	if options.NS {
		count++
	}
	if options.TXT {
		count++
	}
	if options.SRV {
		count++
	}
	if options.PTR {
		count++
	}
	if options.MX {
		count++
	}
	if options.SOA {
		count++
	}
	if options.ANY {
		count++
	}
	if options.AXFR {
		count++
	}
	if options.CAA {
		count++
	}
	return count
}

// estimateCandidates counts input candidates without consuming stdin.
func estimateCandidates(options *Options) int {
	count := 0
	if options.Hosts != "" {
		count += countItems(options.Hosts)
	}
	if options.Domains != "" && options.WordList != "" {
		count += countItems(options.Domains) * countItems(options.WordList)
	}
	return count
}

func countItems(arg string) int {
	if fileutil.FileExists(arg) {
		n := 0
		for line, err := range fileutil.Lines(arg) {
			if err != nil {
				break
			}
			if strings.TrimSpace(line) != "" {
				n++
			}
		}
		return n
	}
	if arg == "" {
		return 0
	}
	// Comma-separated or single item.
	items := strings.Split(arg, ",")
	n := 0
	for _, it := range items {
		if strings.TrimSpace(it) != "" {
			n++
		}
	}
	return n
}

// loadScope reads the scope file and normalizes each line into an allowed
// DNS suffix. One domain per line; comments (#) and blank lines are ignored.
func (options *Options) loadScope() error {
	if !fileutil.FileExists(options.ScopeFile) {
		return fmt.Errorf("scope file not found: %s", options.ScopeFile)
	}
	var scope []string
	for line, err := range fileutil.Lines(options.ScopeFile) {
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		norm := normalizeScopeDomain(line)
		if norm == "" {
			continue
		}
		scope = append(scope, norm)
	}
	options.scope = scope
	options.scopeLoaded = true
	gologger.Print().Msgf("scope loaded: %d allowed suffix(es) from %s", len(scope), options.ScopeFile)
	return nil
}

// normalizeScopeDomain lowercases and strips a trailing root dot from a scope
// entry. Invalid entries return "".
func normalizeScopeDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimSuffix(domain, ".")
	if domain == "" {
		return ""
	}
	// Reject entries containing whitespace, control chars, or wildcards.
	for _, r := range domain {
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '*' {
			return ""
		}
	}
	// Reject IP addresses (scope is domain-suffix based, not IP based).
	if net.ParseIP(domain) != nil {
		return ""
	}
	_, ok := dns.IsDomainName(dns.Fqdn(domain))
	if !ok {
		return ""
	}
	return domain
}

// isInScope reports whether candidate is allowed by the loaded scope. A
// candidate is in scope if it equals a scope entry or is a subdomain of one
// (i.e. ends with "." + scopeEntry). This prevents suffix-spoofing:
// "example.com.attacker.invalid" is NOT in scope of "example.com".
func (options *Options) isInScope(candidate string) bool {
	if !options.scopeLoaded || len(options.scope) == 0 {
		return true // no scope configured: allow everything
	}
	cand := normalizeScopeDomain(candidate)
	if cand == "" {
		return false
	}
	for _, suffix := range options.scope {
		if cand == suffix {
			return true
		}
		if strings.HasSuffix(cand, "."+suffix) {
			return true
		}
	}
	return false
}

func (options *Options) validateOptions() {
	if options.Response && options.ResponseOnly {
		gologger.Fatal().Msgf("resp and resp-only can't be used at the same time")
	}

	if options.OutputTemplate != "" && (options.JSON || options.Raw) {
		gologger.Fatal().Msgf("output-template can't be used with json or raw output")
	}

	if options.Retries == 0 {
		gologger.Fatal().Msgf("retries must be at least 1")
	}

	wordListPresent := options.WordList != ""
	domainsPresent := options.Domains != ""
	hostsPresent := options.Hosts != ""

	if hostsPresent && (wordListPresent || domainsPresent) {
		gologger.Fatal().Msgf("list(l) flag can not be used domain(d) or wordlist(w) flag")
	}

	if wordListPresent && !domainsPresent {
		gologger.Fatal().Msg("missing domain(d) flag required with wordlist(w) input")
	}
	if domainsPresent && !wordListPresent {
		gologger.Fatal().Msgf("missing wordlist(w) flag required with domain(d) input")
	}

	// stdin can be set only on one flag
	if (argumentHasStdin(options.Domains) && argumentHasStdin(options.WordList)) ||
		(argumentHasStdin(options.Domains) && argumentHasStdin(options.Hosts)) ||
		(argumentHasStdin(options.WordList) && argumentHasStdin(options.Hosts)) {
		if options.Stream {
			gologger.Fatal().Msgf("argument stdin not supported in stream mode")
		}
		gologger.Fatal().Msgf("stdin can be set for one flag")
	}

	if options.Stream {
		if wordListPresent {
			gologger.Fatal().Msgf("wordlist not supported in stream mode")
		}
		if domainsPresent {
			gologger.Fatal().Msgf("domains not supported in stream mode")
		}
		if hostsPresent {
			gologger.Fatal().Msgf("hosts not supported in stream mode")
		}
		if options.Resume {
			gologger.Fatal().Msgf("resume not supported in stream mode")
		}
		if options.hasWildcardFiltering() {
			gologger.Fatal().Msgf("wildcard not supported in stream mode")
		}
		if options.ShowStatistics {
			gologger.Fatal().Msgf("stats not supported in stream mode")
		}
	}

	if options.AutoWildcard && options.WildcardDomain != "" {
		gologger.Fatal().Msgf("auto-wildcard and wildcard-domain can't be used at the same time")
	}
}

func (options *Options) hasWildcardFiltering() bool {
	return options.AutoWildcard || options.WildcardDomain != ""
}

func argumentHasStdin(arg string) bool {
	return arg == stdinMarker
}

func normalizeWildcardDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "*.")
	domain = strings.TrimSuffix(domain, ".")
	return domain
}

func normalizeAndValidateWildcardDomain(raw string) (string, error) {
	normalized := normalizeWildcardDomain(raw)
	if raw == "" {
		return normalized, nil
	}
	if !isValidWildcardDomain(normalized) {
		return "", errors.New("invalid wildcard domain")
	}
	return normalized, nil
}

func isValidWildcardDomain(domain string) bool {
	if domain == "" {
		return false
	}
	if strings.Contains(domain, "*") || strings.Contains(domain, "..") || strings.ContainsAny(domain, " \t\r\n") {
		return false
	}
	_, ok := dns.IsDomainName(dns.Fqdn(domain))
	return ok
}

// configureOutput configures the output on the screen
func (options *Options) configureOutput() {
	// If the user desires verbose output, show verbose output
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	if options.NoColor {
		updateutils.Aurora = aurora.New(aurora.WithColors(false))
		gologger.DefaultLogger.SetFormatter(formatter.NewCLI(true))
	}
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

func (options *Options) configureRcodes() error {
	options.rcodes = make(map[int]struct{})
	rcodes := strings.Split(options.RCode, ",")
	for _, rcode := range rcodes {
		var rc int
		switch strings.ToLower(rcode) {
		case "":
			continue
		case "noerror":
			rc = 0
		case "formerr":
			rc = 1
		case "servfail":
			rc = 2
		case "nxdomain":
			rc = 3
		case "notimp":
			rc = 4
		case "refused":
			rc = 5
		case "yxdomain":
			rc = 6
		case "yxrrset":
			rc = 7
		case "nxrrset":
			rc = 8
		case "notauth":
			rc = 9
		case "notzone":
			rc = 10
		case "badsig", "badvers":
			rc = 16
		case "badkey":
			rc = 17
		case "badtime":
			rc = 18
		case "badmode":
			rc = 19
		case "badname":
			rc = 20
		case "badalg":
			rc = 21
		case "badtrunc":
			rc = 22
		case "badcookie":
			rc = 23
		default:
			var err error
			rc, err = strconv.Atoi(rcode)
			if err != nil {
				return errors.New("invalid rcode value")
			}
		}

		options.rcodes[rc] = struct{}{}
	}

	options.hasRCodes = options.RCode != ""
	return nil
}

func (options *Options) configureResume() error {
	options.resumeCfg = &ResumeCfg{}
	if options.Resume && fileutil.FileExists(DefaultResumeFile) {
		return goconfig.Load(&options.resumeCfg, DefaultResumeFile)

	}
	return nil
}

func (options *Options) configureQueryOptions() {
	queryMap := map[string]*bool{
		"a":     &options.A,
		"aaaa":  &options.AAAA,
		"cname": &options.CNAME,
		"ns":    &options.NS,
		"txt":   &options.TXT,
		"srv":   &options.SRV,
		"ptr":   &options.PTR,
		"mx":    &options.MX,
		"soa":   &options.SOA,
		"axfr":  &options.AXFR,
		"caa":   &options.CAA,
		"any":   &options.ANY,
	}

	if options.QueryAll {
		for _, val := range queryMap {
			*val = true
		}
		options.Response = true
		// the ANY query type is not supported by the retryabledns library,
		// thus it's hard to filter the results when it's used in combination with other query types
		options.ExcludeType = append(options.ExcludeType, "any")
	}

	for _, et := range options.ExcludeType {
		if val, ok := queryMap[et]; ok {
			*val = false
		}
	}
}

func (options *Options) configureResolverMode() {
	mode := strings.TrimSpace(strings.ToLower(options.ResolverMode))
	switch mode {
	case "", "first-success":
		options.resolverMode = dnsx.ResolverModeFirstSuccess
	case "all":
		options.resolverMode = dnsx.ResolverModeAll
	case "quorum":
		options.resolverMode = dnsx.ResolverModeQuorum
	case "compare":
		options.resolverMode = dnsx.ResolverModeCompare
	default:
		gologger.Fatal().Msgf("invalid resolver-mode %q: expected first-success, all, quorum or compare", options.ResolverMode)
	}
}
