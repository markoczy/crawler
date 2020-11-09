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
	"github.com/markoczy/crawler/perm"
	"github.com/markoczy/crawler/types"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	cfg := cli.ParseFlags()
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getAllLinks(cfg, ctx)
	for _, link := range links.Values() {
		if cfg.Download() {
			fmt.Printf("Downloading from URL '%s'\n", link)
			if err := httpfunc.DownloadFile(cfg, link); err != nil {
				fmt.Printf("ERROR: Failed to download content at url '%s': %s\n", link, err.Error())
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
	perms := perm.Perm(cfg.Url())
	allLinks := types.NewStringSet()
	allLinks.Add(perms...)
	allVisited := types.NewStringSet()
	for _, perm := range perms {
		links := getLinksRecursive(cfg, ctx, perm, 0, allVisited)
		for _, link := range links.Values() {
			b := []byte(link)
			if !cfg.Include().Match(b) || cfg.Exclude().Match(b) {
				links.Remove(link)
			}
		}
		allLinks.Add(links.Values()...)
	}
	return allLinks
}

func getLinksRecursive(cfg cli.CrawlerConfig, ctx context.Context, url string, depth int, visited *types.StringSet) *types.StringSet {
	// exit condition 1: over depth (download mode has depth-1)
	if depth > cfg.Depth() || (cfg.Download() && depth > cfg.Depth()-1) {
		return types.NewStringSet()
	}
	// exit condition 2: already visited
	if visited.Exists(url) {
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
	visited.Add(url)
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
		b := []byte(val)
		if cfg.FollowInclude().Match(b) && !cfg.FollowExclude().Match(b) {
			ret = append(ret, val)
		}
	}
	return ret, nil
}
