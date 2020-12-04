package cli

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/markoczy/crawler/perm"
)

const (
	errUndefinedFlag = 100
	errParseFailed   = 200
	errGeneral       = 500
	unset            = "<unset>"
	none             = "none"
	empty            = ""
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36"
	matchAll         = ".*"
	matchNothing     = "$^"
)

func ParseFlags() CrawlerConfig {
	var err error
	var headerFlags arrayValue
	cfg := crawlerConfig{}

	testPtr := flag.Bool("test", false, "tests patterns and outputs download file name")
	urlPtr := flag.String("url", unset, "the initial url, cannot be unset, prefix http or https is required, supports permutations in square brackets like '[1-100]' or '[a,b,c]', can also refer a file with prefix '@'")
	downloadPtr := flag.Bool("download", false, "switches to download mode")
	timeoutPtr := flag.Int64("timeout", 60000, "general timeout in millis when loading a webpage")
	extraWaittimePtr := flag.Int64("extra-waittime", 0, "additional waittime after load")
	depthPtr := flag.Int("depth", 0, "max depth for link crawler")
	flag.Var(&headerFlags, "header", "headers to set, multiple allowed, prefix '@' to adress a file")
	authPtr := flag.String("auth", unset, "basic auth header to set, auth must be provided in format 'user:password'")
	userAgentPtr := flag.String("user-agent", unset, "user agent to set, defaults to chrome browser if unset, set 'none' to avoid overriding user agent")
	includePtr := flag.String("include", matchAll, "regex of included links, defaults to 'match all' (hint: prefix '(?flags)' to define flags)")
	excludePtr := flag.String("exclude", matchNothing, "regex of excluded links, defaults to 'match nothing' (hint: prefix '(?flags)' to define flags)")
	followIncludePtr := flag.String("follow-include", matchAll, "regex of included links to follow, only applies if depth>0, defaults to 'match all' (hint: prefix '(?flags)' to define flags)")
	followExcludePtr := flag.String("follow-exclude", matchNothing, "regex of excluded links to follow, only applies if depth>0, defaults to 'match nothing' (hint: prefix '(?flags)' to define flags)")
	logFilePtr := flag.String("logfile", unset, "path to log file, defaults to stdout when unset")
	namingCapturePtr := flag.String("naming-capture", `^http(s|)://(?P<path>.*)/(?P<name>\w+)(\.|)(?P<ext>(\.\w+)|)$`, "regex for capturing groups of output file name, use in combination with 'naming-pattern', only applies to download mode")
	namingCaptureFoldersPtr := flag.Bool("naming-capture-folders", false, "specifies wether '/' inside capture groups are treated as subfolders, if false the '/' characters in the capture groups are replaced by '_', only applies to download mode")
	namingPatternPtr := flag.String("naming-pattern", "<path>/<name><ext>", "pattern to resolve output file name, use '<name>' to reference a capture group from 'naming-capture' flag, only applies to download mode")
	reconnectAttemptsPtr := flag.Int("reconnect", 5, "Amount of reconnect attempts when context was closed")
	flag.Parse()

	url := *urlPtr
	cfg.test = *testPtr
	cfg.download = *downloadPtr
	cfg.depth = *depthPtr
	cfg.include = parseRegex(*includePtr, "include")
	cfg.exclude = parseRegex(*excludePtr, "exclude")
	cfg.followInclude = parseRegex(*followIncludePtr, "follow-include")
	cfg.followExclude = parseRegex(*followExcludePtr, "follow-exclude")
	cfg.namingCapture = parseRegex(*namingCapturePtr, "naming-pattern")
	cfg.namingCaptureFolders = *namingCaptureFoldersPtr
	cfg.namingPattern = *namingPatternPtr
	cfg.reconnectAttempts = *reconnectAttemptsPtr

	cfg.timeout = time.Duration(*timeoutPtr) * time.Millisecond
	cfg.extraWaittime = time.Duration(*extraWaittimePtr) * time.Millisecond
	logFile := *logFilePtr
	if logFile != unset {
		var file *os.File
		if file, err = os.Create(logFile); err != nil {
			exitError("Failed to create log file "+logFile, errGeneral)
		}
		log.SetOutput(file)
	}
	if url == unset {
		exitError(fmt.Sprintf("Mandatory value 'url' was not defined"), errUndefinedFlag)
	}
	cfg.urls = parseUrls(url)
	if cfg.headers, err = parseHeaderFlags(headerFlags.Values()); err != nil {
		exitError(fmt.Sprintf("Parse of value 'header' failed: %s", err.Error()), errParseFailed)
	}
	auth := *authPtr
	if auth != unset {
		addAuthHeader(auth, &cfg.headers)
	}
	userAgent := *userAgentPtr
	if userAgent == unset {
		addUserAgentHeader(defaultUserAgent, &cfg.headers)
	} else if strings.ToLower(userAgent) != none {
		addUserAgentHeader(userAgent, &cfg.headers)
	}

	return &cfg
}

func exitError(s string, code int) {
	flag.Usage()
	fmt.Println("\nERROR: " + s)
	os.Exit(code)
}

func parseHeaderFlags(headerFlags []string) (map[string]string, error) {
	var err error
	ret := map[string]string{}
	for _, s := range headerFlags {
		if strings.HasPrefix(s, "@") {
			if err = loadHeaderFile(s[1:], &ret); err != nil {
				return nil, err
			}
			continue
		}
		if err = parseHeaderFlag(s, &ret); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func loadHeaderFile(path string, m *map[string]string) error {
	var err error
	var dat []byte
	if dat, err = ioutil.ReadFile(path); err != nil {
		return err
	}
	split := strings.Split(string(dat), "\n")
	for _, s := range split {
		if err = parseHeaderFlag(s, m); err != nil {
			return err
		}
	}
	return nil
}

func parseHeaderFlag(token string, m *map[string]string) error {
	token = strings.TrimSpace(token)
	if token == empty {
		return nil
	}
	split := strings.SplitN(token, ":", 2)
	if len(split) < 2 {
		return fmt.Errorf("Could not parse header for token '%s' missing key value separator ':'", token)
	}
	key := strings.TrimSpace(split[0])
	val := strings.TrimSpace(split[1])
	(*m)[key] = val
	return nil
}

func addAuthHeader(auth string, m *map[string]string) {
	val := base64.StdEncoding.EncodeToString([]byte(auth))
	(*m)["authorization"] = "Basic " + val
}

func addUserAgentHeader(userAgent string, m *map[string]string) {
	(*m)["user-agent"] = userAgent
}

func parseRegex(val, name string) *regexp.Regexp {
	ret, err := regexp.Compile(val)
	if err != nil {
		exitError(fmt.Sprintf("Regex '%s' could not be compiled: %s", name, err.Error()), errParseFailed)
	}
	return ret
}

func parseUrls(url string) []string {
	var dat []byte
	var err error
	ret := []string{}
	if strings.HasPrefix(url, "@") {
		if dat, err = ioutil.ReadFile(url[1:]); err != nil {
			exitError(fmt.Sprintf("Could not read file '%s'", url[1:]), errParseFailed)
		}
		split := strings.Split(string(dat), "\n")
		for _, v := range split {
			cur := strings.TrimSpace(v)
			if cur != empty {
				ret = append(ret, perm.Perm(cur)...)
			}
		}
	} else {
		ret = append(ret, perm.Perm(url)...)
	}
	return ret
}

type arrayValue []string

func (i *arrayValue) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayValue) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

func (i *arrayValue) Values() []string {
	return []string(*i)
}
