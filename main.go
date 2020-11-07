package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/actions"
	"github.com/markoczy/crawler/types"
	"golang.org/x/exp/errors/fmt"
)

const (
	errUndefinedFlag = 100
	unset            = "<unset>"
)

var (
	url     string
	depth   int
	timeout time.Duration
)

func parseFlags() {
	urlPtr := flag.String("url", unset, "the initial url (prefix http or https needed)")
	timeoutPtr := flag.Int64("timeout", 10000, "general timeout in millis when loading a webpage")
	depthPtr := flag.Int("depth", 0, "max depth for link crawler")

	flag.Parse()
	url, depth = *urlPtr, *depthPtr
	timeout = time.Duration(*timeoutPtr) * time.Millisecond
	if url == unset {
		exitErrUndefined("url")
	}
}

func exitErrUndefined(val string) {
	flag.Usage()
	fmt.Printf("\nERROR: Mandatory value '%s' was not defined\n", val)
	os.Exit(errUndefinedFlag)
}

func main() {
	parseFlags()
	execGetLinks()
}

func execGetLinks() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	links := getLinksRecursive(ctx, url, timeout, 0, depth, types.NewStringSet())
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

func getScript(name string) string {
	dat, err := ioutil.ReadFile("js/" + name)
	check(err)
	return string(dat)
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
		chromedp.Evaluate(getScript("getLinks.js"), &ret),
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
		// if at max depth, the site was not crawled...
		if depth < maxDepth {
			visited.Add(link)
		}
	}
	return ret
}
