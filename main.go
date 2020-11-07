package main

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/actions"
	"github.com/markoczy/crawler/cli"
	"github.com/markoczy/crawler/js"
	"github.com/markoczy/crawler/types"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	cfg := cli.ParseFlags()
	fmt.Println("Headers:", cfg.Headers())
}

func execGetLinks(cfg cli.CrawlerConfig) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getLinksRecursive(ctx, cfg.Url(), cfg.Timeout(), 0, cfg.Depth(), types.NewStringSet())
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

func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}

func getLinks(ctx context.Context, url string, timeout time.Duration) ([]string, error) {
	var ret []string
	tasks := chromedp.Tasks{
		actions.NavigateAndWaitLoaded(url, timeout),
		chromedp.Evaluate(js.GetLinks, &ret),
	}
	if err := chromedp.Run(ctx, tasks); err != nil {
		return ret, err
	}
	return ret, nil
}

func getLinksRecursive(ctx context.Context, url string, timeout time.Duration, depth, maxDepth int, visited *types.StringSet) *types.StringSet {
	// exit condition 1: over depth
	if depth > maxDepth {
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
	if links, err = getLinks(ctx, url, timeout); err != nil {
		log.Printf("ERROR: Failed to get links from url '%s': %s\n", url, err.Error())
	} else {
		log.Printf("Found %d links at url '%s'\n", len(links), url)
	}
	visited.Add(url)
	ret := types.NewStringSet()
	ret.Add(links...)

	for _, link := range links {
		more := getLinksRecursive(ctx, link, timeout, depth+1, maxDepth, visited)
		ret.Add(more.Values()...)
	}
	return ret
}
