// Command screenshot is a chromedp example demonstrating how to take a
// screenshot of a specific element and of the entire browser viewport.
package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/action"
	"golang.org/x/exp/errors/fmt"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var links []string

	if err := chromedp.Run(ctx, getLinks(`https://www.google.ch/`, &links)); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Links", links)

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

func getLinks(url string, ret *[]string) chromedp.Tasks {
	return chromedp.Tasks{
		action.NavigateAndWaitLoaded(url, 10*time.Second),
		chromedp.Evaluate(getScript("getLinks.js"), ret),
	}
}
