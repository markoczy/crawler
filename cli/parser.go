package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	errUndefinedFlag = 100
	errParseFailed   = 200
	unset            = "<unset>"
)

func ParseFlags() CrawlerConfig {
	var err error
	var headerFlags arrayValue
	cfg := crawlerConfig{}

	urlPtr := flag.String("url", unset, "the initial url (prefix http or https needed)")
	timeoutPtr := flag.Int64("timeout", 10000, "general timeout in millis when loading a webpage")
	depthPtr := flag.Int("depth", 0, "max depth for link crawler")
	flag.Var(&headerFlags, "header", "headers to set, multiple allowed, prefix '@' to adress a file")
	userPtr := flag.String("user", unset, "basic auth user")
	passPtr := flag.String("pass", unset, "basic auth password (mandatory when 'user' is set)")
	flag.Parse()

	cfg.url, cfg.depth = *urlPtr, *depthPtr
	cfg.timeout = time.Duration(*timeoutPtr) * time.Millisecond
	if cfg.url == unset {
		exitError(fmt.Sprintf("Mandatory value 'url' was not defined"), errUndefinedFlag)
	}
	if cfg.headers, err = parseHeaderFlags(headerFlags.Values()); err != nil {
		exitError(fmt.Sprintf("Parse of value 'header' failed: %s", err.Error()), errParseFailed)
	}
	user, pass := *userPtr, *passPtr
	if user != unset && pass == unset {
		exitError("Password must be set when user is set", errUndefinedFlag)
	}

	return &cfg
}

func exitError(s string, code int) {
	flag.Usage()
	fmt.Println("\nERROR:" + s)
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
	if token == "" {
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
