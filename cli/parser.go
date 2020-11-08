package cli

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	errUndefinedFlag = 100
	errParseFailed   = 200
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

	urlPtr := flag.String("url", unset, "the initial url,prefix http or https needed, supports permutations in square brackets like '[1-100]' or '[a,b,c]'")
	timeoutPtr := flag.Int64("timeout", 10000, "general timeout in millis when loading a webpage")
	depthPtr := flag.Int("depth", 0, "max depth for link crawler")
	flag.Var(&headerFlags, "header", "headers to set, multiple allowed, prefix '@' to adress a file")
	authPtr := flag.String("auth", unset, "basic auth header to set, auth must be provided in format 'user:password'")
	userAgentPtr := flag.String("user-agent", unset, "user agent to set, defaults to chrome browser if unset, set 'none' to avoid overriding user agent")
	includePtr := flag.String("include", matchAll, "regex of included links, defaults to 'match all' (hint: prefix '(?flags)' to define flags)")
	excludePtr := flag.String("exclude", matchNothing, "regex of excluded links, defaults to 'match nothing' (hint: prefix '(?flags)' to define flags)")
	followIncludePtr := flag.String("follow-include", matchAll, "regex of included links to follow, only applies if depth>0, defaults to 'match all' (hint: prefix '(?flags)' to define flags)")
	followExcludePtr := flag.String("follow-exclude", matchNothing, "regex of excluded links to follow, only applies if depth>0, defaults to 'match nothing' (hint: prefix '(?flags)' to define flags)")
	flag.Parse()

	cfg.include = parseRegex(*includePtr, "include")
	cfg.exclude = parseRegex(*excludePtr, "exclude")
	cfg.followInclude = parseRegex(*followIncludePtr, "follow-include")
	cfg.followExclude = parseRegex(*followExcludePtr, "follow-exclude")

	cfg.url, cfg.depth = *urlPtr, *depthPtr
	cfg.timeout = time.Duration(*timeoutPtr) * time.Millisecond

	if cfg.url == unset {
		exitError(fmt.Sprintf("Mandatory value 'url' was not defined"), errUndefinedFlag)
	}
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

func parseHeaderFlags(headerFlags []string) (map[string]interface{}, error) {
	var err error
	ret := map[string]interface{}{}
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

func loadHeaderFile(path string, m *map[string]interface{}) error {
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

func parseHeaderFlag(token string, m *map[string]interface{}) error {
	token = strings.TrimSpace(token)
	if token == empty {
		return nil
	}
	split := strings.SplitN(token, ":", 2)
	if len(split) < 2 {
		return fmt.Errorf("Could not parse header for token '%s' missing key value separator ':'", token)
	}
	// avoid dupes by setting lowercase (header keys are case insensitive)
	key := strings.ToLower(strings.TrimSpace(split[0]))
	val := strings.TrimSpace(split[1])
	(*m)[key] = val
	return nil
}

func addAuthHeader(auth string, m *map[string]interface{}) {
	val := base64.StdEncoding.EncodeToString([]byte(auth))
	(*m)["authorization"] = "Basic " + val
}

func addUserAgentHeader(userAgent string, m *map[string]interface{}) {
	(*m)["user-agent"] = userAgent
}

func parseRegex(val, name string) *regexp.Regexp {
	ret, err := regexp.Compile(val)
	if err != nil {
		exitError(fmt.Sprintf("Regex '%s' could not be compiled: %s", name, err.Error()), errParseFailed)
	}
	return ret
}
