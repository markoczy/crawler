package main

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/actions"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/httpfunc"
	"github.com/markoczy/crawler/js"
	"github.com/markoczy/crawler/types"
	"golang.org/x/exp/errors/fmt"
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
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getAllLinks(cfg, ctx)
	for _, link := range links.Values() {
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

func getAllLinks(cfg cli.CrawlerConfig, ctx context.Context) *types.StringSet {
	allLinks := types.NewStringSet()
	for _, url := range cfg.Urls() {
		if !cfg.Include().MatchString(url) || cfg.Exclude().MatchString(url) {
			fmt.Printf("Not including '%s': URL not matching include or matching exclude pattern\n", url)
			continue
		}
		allLinks.Add(url)
	}
	allVisited := types.NewTracker()
	for _, perm := range cfg.Urls() {
		links := getLinksRecursive(cfg, ctx, perm, 0, allVisited)
		for _, link := range links.Values() {
			if !cfg.Include().MatchString(link) || cfg.Exclude().MatchString(link) {
				fmt.Printf("Not including '%s': URL not matching include or matching exclude pattern\n", link)
				links.Remove(link)
			}
		}
		allLinks.Add(links.Values()...)
	}
	return allLinks
}

func getLinksRecursive(cfg cli.CrawlerConfig, ctx context.Context, url string, depth int, visited *types.Tracker) *types.StringSet {
	// exit condition 1: over depth (download mode has depth-1)
	if depth > cfg.Depth() || (cfg.Download() && depth > cfg.Depth()-1) {
		return types.NewStringSet()
	}
	// exit condition 2: already visited
	if !visited.ShouldVisit(url, depth) {
		log.Printf("Already visited '%s'\n", url)
		return types.NewStringSet()
	}

	log.Printf("Scanning url '%s'", url)
	var links []string
	var err error
	if links, err = getLinks(cfg, ctx, url); err != nil {
		log.Printf("ERROR: Failed to get links from url '%s': %s\n", url, err.Error())
	} else {
		log.Printf("Found %d links at url '%s'\n", len(links), url)
	}
	visited.Add(url, depth)
	ret := types.NewStringSet()
	ret.Add(links...)

	for _, link := range links {
		more := getLinksRecursive(cfg, ctx, link, depth+1, visited)
		ret.Add(more.Values()...)
	}
	return ret
}

func getLinks(cfg cli.CrawlerConfig, ctx context.Context, url string) ([]string, error) {
	var buf []string
	tasks := chromedp.Tasks{}
	if len(cfg.Headers()) > 0 {
		tasks = append(tasks, network.SetExtraHTTPHeaders(network.Headers(cfg.Headers())))
	}
	tasks = append(tasks,
		actions.NavigateAndWaitLoaded(url, cfg.Timeout()),
		chromedp.Evaluate(js.GetLinks, &buf),
	)
	if err := chromedp.Run(ctx, tasks); err != nil {
		return []string{}, err
	}
	ret := []string{}
	for _, val := range buf {
		if cfg.FollowInclude().MatchString(val) && !cfg.FollowExclude().MatchString(val) {
			fmt.Printf("Not following link '%s': URL not matching follow-include or matching follow-exclude pattern\n", val)
			ret = append(ret, val)
		}
	}
	return ret, nil
}
