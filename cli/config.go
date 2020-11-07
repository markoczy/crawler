package cli

import (
	"fmt"
	"strings"
	"time"
)

type CrawlerConfig interface {
	Url() string
	Depth() int
	Timeout() time.Duration
	Headers() map[string]interface{}
}

type crawlerConfig struct {
	url     string
	depth   int
	timeout time.Duration
	headers map[string]interface{}
}

func (cfg *crawlerConfig) Url() string {
	return cfg.url
}

func (cfg *crawlerConfig) Depth() int {
	return cfg.depth
}

func (cfg *crawlerConfig) Timeout() time.Duration {
	return cfg.timeout
}

func (cfg *crawlerConfig) Headers() map[string]interface{} {
	return cfg.headers
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
