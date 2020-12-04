package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	// "context"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/httpfunc"
	"github.com/markoczy/crawler/js"
	"github.com/markoczy/crawler/types"
)

var (
	browser *rod.Browser
	router  *rod.HijackRouter
)

func main() {
	cfg := cli.ParseFlags()
	log.Println("Parsed Params:", cfg)
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
			log.Printf("Downloading from URL '%s'\n", link)
			if err := httpfunc.DownloadFile(cfg, link); err != nil {
				log.Printf("ERROR: Failed to download content at url '%s': %s\n", link, err.Error())
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
				log.Printf("Not including '%s': URL not matching include or matching exclude pattern\n", link)
				links.Remove(link)
				continue
			}
			log.Printf("Found Link '%s'\n", link)
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
		log.Printf("Already visited '%s'\n", url)
		return ret
	}

	log.Printf("Scanning url '%s'", url)
	var links []string
	var err error
	if links, err = getLinks(cfg, url); err != nil {
		if err.Error() == context.Canceled.Error() {
			log.Printf("WARN: Failed to get links from url '%s': Context was canceled, retrying...\n", url)
			retryAttempts := 1
			shouldRetry := true
			success := false
			for retryAttempts <= cfg.ReconnectAttempts() && shouldRetry {
				log.Printf("Retry attempt %d of %d\n", retryAttempts, cfg.ReconnectAttempts())
				reconnect(cfg)
				if links, err = getLinks(cfg, url); err != nil {
					shouldRetry = err.Error() == context.Canceled.Error()
				} else {
					log.Printf("Succeeded at retry attempt %d\n", retryAttempts)
					shouldRetry = false
					success = true
				}
				retryAttempts++
			}
			if !success {
				log.Printf("ERROR: Failed to get links from url '%s': %s\n", url, err.Error())
			}
		} else {
			log.Printf("ERROR: Failed to get links from url '%s': %s\n", url, err.Error())
		}
	} else {
		log.Printf("Found %d links at url '%s'\n", len(links), url)
	}
	ret.Add(links...)
	visited.Add(url, depth)

	for _, link := range links {
		if !cfg.FollowInclude().MatchString(link) || cfg.FollowExclude().MatchString(link) {
			log.Printf("Not following link '%s': URL not matching follow-include or matching follow-exclude pattern\n", link)
			continue
		}
		log.Printf("Following link '%s'\n", link)
		more := getLinksRecursive(cfg, link, depth+1, visited)
		ret.Add(more.Values()...)
	}
	return ret
}

func getLinks(cfg cli.CrawlerConfig, url string) ([]string, error) {
	var err error
	var resp *proto.RuntimeRemoteObject
	var page *rod.Page
	ret := []string{}
	// Navigate and load
	if page, err = browser.Page(proto.TargetCreateTarget{}); err != nil {
		return ret, err
	}
	// Set Headers
	var cleanup func()
	if err = page.Navigate(url); err != nil {
		return ret, nil
	}
	pageWithTimeout := page.Timeout(cfg.Timeout())
	if err = pageWithTimeout.WaitLoad(); err != nil {
		return ret, err
	}
	// Wait additional time
	if cfg.ExtraWaittime() != 0 {
		if _, err = page.Evaluate(js.CreateWaitFunc(cfg.ExtraWaittime())); err != nil {
			return ret, err
		}
	}
	// Get links
	if resp, err = page.Eval(js.GetLinks); err != nil {
		return ret, err
	}
	for _, link := range resp.Value.Arr() {
		ret = append(ret, link.String())
	}
	// Cleanup Context and close tab
	if cleanup != nil {
		cleanup()
	}
	err = page.Close()
	return ret, err
}

func reconnect(cfg cli.CrawlerConfig) {
	disconnect()
	browser = rod.New().MustConnect()
	router = browser.HijackRequests()
	router.MustAdd("*/*", func(ctx *rod.Hijack) {
		for k, v := range cfg.Headers() {
			ctx.Request.Req().Header.Set(k, v)
		}
		ctx.LoadResponse(http.DefaultClient, true)
	})
	go router.Run()
}

func disconnect() {
	if router != nil {
		router.Stop()
	}
	if browser != nil {
		browser.Close()
	}
}
