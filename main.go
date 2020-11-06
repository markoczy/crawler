package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/action"
	"golang.org/x/exp/errors/fmt"
)

const (
	errUndefinedFlag = 100
	unset            = "<unset>"
)

var (
	url string
)

func parseFlags() {
	urlPtr := flag.String("url", unset, "the url")

	flag.Parse()
	url = *urlPtr
	if url == unset {
		exitErrUndefined("url")
	}
}

func exitErrUndefined(val string) {
	flag.Usage()
	fmt.Printf("\nMandatory value '%s' was not defined\n", val)
	os.Exit(errUndefinedFlag)
}

func main() {
	parseFlags()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var links []string
	var err error

	if links, err = getLinks(ctx, url, 10*time.Second); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Links", links)
	chromedp.Cancel(ctx)
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
		action.NavigateAndWaitLoaded(url, timeout),
		chromedp.Evaluate(getScript("getLinks.js"), &ret),
	}
	if err := chromedp.Run(ctx, tasks); err != nil {
		return nil, err
	}
	return ret, nil
}
