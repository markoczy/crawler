package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type CrawlerConfig interface {
	Url() string
	Depth() int
	Timeout() time.Duration
	Headers() map[string]interface{}
	Include() *regexp.Regexp
	Exclude() *regexp.Regexp
	FollowInclude() *regexp.Regexp
	FollowExclude() *regexp.Regexp
}

type crawlerConfig struct {
	url           string
	depth         int
	timeout       time.Duration
	headers       map[string]interface{}
	include       *regexp.Regexp
	exclude       *regexp.Regexp
	followInclude *regexp.Regexp
	followExclude *regexp.Regexp
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

func (cfg *crawlerConfig) Include() *regexp.Regexp {
	return cfg.include
}

func (cfg *crawlerConfig) Exclude() *regexp.Regexp {
	return cfg.exclude
}

func (cfg *crawlerConfig) FollowInclude() *regexp.Regexp {
	return cfg.followInclude
}

func (cfg *crawlerConfig) FollowExclude() *regexp.Regexp {
	return cfg.followExclude
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
