package cli

import (
	"fmt"
	"regexp"
	"time"
)

type CrawlerConfig interface {
	Test() bool
	Urls() []string
	Download() bool
	Depth() int
	Timeout() time.Duration
	ExtraWaittime() time.Duration
	Headers() map[string]interface{}
	Include() *regexp.Regexp
	Exclude() *regexp.Regexp
	FollowInclude() *regexp.Regexp
	FollowExclude() *regexp.Regexp
	NamingCapture() *regexp.Regexp
	NamingCaptureFolders() bool
	NamingPattern() string
	String() string
}

type crawlerConfig struct {
	test                 bool
	urls                 []string
	download             bool
	depth                int
	timeout              time.Duration
	extraWaittime        time.Duration
	headers              map[string]interface{}
	include              *regexp.Regexp
	exclude              *regexp.Regexp
	followInclude        *regexp.Regexp
	followExclude        *regexp.Regexp
	namingCapture        *regexp.Regexp
	namingCaptureFolders bool
	namingPattern        string
}

func (cfg *crawlerConfig) Test() bool {
	return cfg.test
}

func (cfg *crawlerConfig) Urls() []string {
	return cfg.urls
}

func (cfg *crawlerConfig) Download() bool {
	return cfg.download
}

func (cfg *crawlerConfig) Depth() int {
	return cfg.depth
}

func (cfg *crawlerConfig) Timeout() time.Duration {
	return cfg.timeout
}

func (cfg *crawlerConfig) ExtraWaittime() time.Duration {
	return cfg.extraWaittime
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

func (cfg *crawlerConfig) NamingCapture() *regexp.Regexp {
	return cfg.namingCapture
}

func (cfg *crawlerConfig) NamingCaptureFolders() bool {
	return cfg.namingCaptureFolders
}

func (cfg *crawlerConfig) NamingPattern() string {
	return cfg.namingPattern
}

func (cfg *crawlerConfig) String() string {
	return fmt.Sprintf("CrawlerConfig [test: '%v', urls: '%v', download: '%v', depth: '%v', timeout: '%v', headers: '%v', include: '%v', exclude: '%v', follow-include: '%v', follow-exclude: '%v', namingCapture: '%v', namingCaptureFolders: '%v', namingPattern: '%v']", cfg.test, cfg.urls, cfg.download, cfg.depth, cfg.timeout, cfg.headers, cfg.include.String(), cfg.exclude.String(), cfg.followInclude.String(), cfg.followExclude.String(), cfg.namingCapture.String(), cfg.namingCaptureFolders, cfg.namingPattern)
}
