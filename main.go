package main

import (
	"context"
	"log"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/actions"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/js"
	"github.com/markoczy/crawler/types"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	cfg := cli.ParseFlags()
	execGetLinks(cfg)
	// fmt.Println("Headers:", cfg.Headers())
}

func execGetLinks(cfg cli.CrawlerConfig) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getLinksRecursive(cfg, ctx, cfg.Url(), 0, types.NewStringSet())
	for _, link := range links.Values() {
		fmt.Println(link)
	}
}

// Helpers

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Maybe outsource

func getLinks(cfg cli.CrawlerConfig, ctx context.Context, url string) ([]string, error) {
	var ret []string
	tasks := chromedp.Tasks{}
	if len(cfg.Headers()) > 0 {
		tasks = append(tasks, network.SetExtraHTTPHeaders(network.Headers(cfg.Headers())))
	}
	tasks = append(tasks,
		actions.NavigateAndWaitLoaded(url, cfg.Timeout()),
		chromedp.Evaluate(js.GetLinks, &ret),
	)
	if err := chromedp.Run(ctx, tasks); err != nil {
		return ret, err
	}
	return ret, nil
}

func getLinksRecursive(cfg cli.CrawlerConfig, ctx context.Context, url string, depth int, visited *types.StringSet) *types.StringSet {
	// exit condition 1: over depth
	if depth > cfg.Depth() {
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
