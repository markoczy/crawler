package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"

	// "context"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/httpfunc"
	"github.com/markoczy/crawler/js"
	"github.com/markoczy/crawler/logger"
	"github.com/markoczy/crawler/types"
)

var (
	browser *rod.Browser
	router  *rod.HijackRouter
	log     logger.Logger

	validConnectErrs = []string{
		"unsupported protocol scheme",
		"no data of the requested type was found",
		"context canceled",
	}
)

func main() {
	cfg := cli.ParseFlags()
	if log == nil {
		// logger may be initialized before (test scope)
		log = logger.New(cfg.LogWarn(), cfg.LogInfo(), cfg.LogDebug())
	}
	log.Info("Parsed Params: %s", cfg.String())
	if cfg.Test() {
		test(cfg)
		return
	}
	exec(cfg)
}

func test(cfg cli.CrawlerConfig) {
	// TODO
}

func exec(cfg cli.CrawlerConfig) {
	reconnect(cfg)
	defer disconnect()

	links := getAllLinks(cfg).Values()
	sort.Strings(links)
	for _, link := range links {
		if cfg.Download() {
			log.Info("Downloading from URL '%s'", link)
			if err := httpfunc.DownloadFile(cfg, link); err != nil {
				log.Error("Failed to download content at url '%s': %s", link, err.Error())
			}
		} else {
			fmt.Println(link)
		}
	}
}

// Helpers

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Maybe outsource

func getAllLinks(cfg cli.CrawlerConfig) *types.StringSet {
	allLinks := types.NewStringSet()
	allVisited := types.NewTracker()
	for _, perm := range cfg.Urls() {
		links := getLinksRecursive(cfg, perm, 0, allVisited)
		for _, link := range links.Values() {
			if !cfg.Include().MatchString(link) || cfg.Exclude().MatchString(link) {
				log.Info("Not including '%s': URL not matching include or matching exclude pattern", link)
				links.Remove(link)
				continue
			}
			log.Info("Found Link '%s'", link)
		}
		allLinks.Add(links.Values()...)
	}
	return allLinks
}

func getLinksRecursive(cfg cli.CrawlerConfig, url string, depth int, visited *types.Tracker) *types.StringSet {
	ret := types.NewStringSet()
	ret.Add(url)
	// exit condition 1: over depth (download mode has depth-1)
	if depth > cfg.Depth() || (cfg.Download() && depth > cfg.Depth()-1) {
		return ret
	}
	// exit condition 2: already visited
	if !visited.ShouldVisit(url, depth) {
		log.Info("Already visited '%s'", url)
		return ret
	}

	log.Info("Scanning url '%s'", url)
	var links []string
	var err error
	if links, err = getLinks(cfg, url); err != nil {
		if err.Error() == context.Canceled.Error() {
			log.Warn("Failed to get links from url '%s': Context was canceled, retrying...", url)
			retryAttempts := 1
			shouldRetry := true
			success := false
			for retryAttempts <= cfg.ReconnectAttempts() && shouldRetry {
				log.Info("Retry attempt %d of %d", retryAttempts, cfg.ReconnectAttempts())
				reconnect(cfg)
				if links, err = getLinks(cfg, url); err != nil {
					shouldRetry = err.Error() == context.Canceled.Error()
				} else {
					log.Info("Succeeded at retry attempt %d", retryAttempts)
					shouldRetry = false
					success = true
				}
				retryAttempts++
			}
			if !success {
				log.Error("Failed to get links from url '%s': %s", url, err.Error())
			}
		} else {
			log.Error("Failed to get links from url '%s': %s", url, err.Error())
		}
	} else {
		log.Info("Found %d links at url '%s'", len(links), url)
	}
	ret.Add(links...)
	visited.Add(url, depth)

	for _, link := range links {
		if !cfg.FollowInclude().MatchString(link) || cfg.FollowExclude().MatchString(link) {
			log.Info("Not following link '%s': URL not matching follow-include or matching follow-exclude pattern\n", link)
			continue
		}
		log.Info("Following link '%s'", link)
		more := getLinksRecursive(cfg, link, depth+1, visited)
		ret.Add(more.Values()...)
	}
	return ret
}

func getLinks(cfg cli.CrawlerConfig, url string) (ret []string, err error) {
	var resp gson.JSON
	var page *rod.Page
	ret = []string{}
	defer func() {
		ex := recover()
		err = getErr(ex)
		if err != nil {
			log.Debug("Cached error at getLinks: %s", err.Error())
		}
		if page != nil {
			log.Debug("Closing page")
			if e2 := page.Close(); e2 != nil {
				log.Debug("Failed to close page: %s", e2.Error())
			}
		}
	}()

	// Navigate and load
	log.Debug("Opening page")
	page = browser.MustPage("")
	log.Debug("Navigating")
	page.Timeout(cfg.Timeout()).MustNavigate(url).MustWaitLoad()

	// Wait additional time
	if cfg.ExtraWaittime() != 0 {
		log.Debug("Waiting for additional waittime")
		page.MustEvaluate(js.CreateWaitFunc(cfg.ExtraWaittime()))
	}

	// Get links
	log.Debug("Running getLinks JS func")
	resp = page.MustEval(js.GetLinks)
	log.Debug("Parsing JSON")
	for _, link := range resp.Arr() {
		ret = append(ret, link.String())
	}
	return
}

func reconnect(cfg cli.CrawlerConfig) {
	disconnect()
	log.Debug("Opening Browser")
	browser = rod.New().MustConnect()
	log.Debug("Adding Hijack Router")
	router = browser.HijackRequests()
	router.MustAdd("*/*", func(ctx *rod.Hijack) {
		headers := []*proto.FetchHeaderEntry{}
		for k, v := range ctx.Request.Headers() {
			headers = append(headers, &proto.FetchHeaderEntry{
				Name:  k,
				Value: v.Str(),
			})
		}
		for k, v := range cfg.Headers() {
			headers = append(headers, &proto.FetchHeaderEntry{
				Name:  k,
				Value: v,
			})
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{Headers: headers})
	})
	go router.Run()
}

func disconnect() {
	if router != nil {
		router.Stop()
	}
	if browser != nil {
		if err := browser.Close(); err != nil {
			log.Debug("Failed to close browser: %s", err.Error())
		}
	}
}

func checkConnectError(err error) bool {
	for _, e := range validConnectErrs {
		strings.Contains(err.Error(), e)
		return true
	}
	return false
}

func getErr(ex interface{}) error {
	if ex != nil {
		switch x := ex.(type) {
		case string:
			return errors.New(x)
		case error:
			return x
		default:
			return errors.New(fmt.Sprintf("%v", x))
		}
	}
	return nil
}
