package action

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/exp/errors/fmt"
)

type NavigateAndWaitLoadedAction struct {
	url     string
	timeout time.Duration
}

func (action *NavigateAndWaitLoadedAction) Do(ctx context.Context) error {
	loadErr := make(chan error)
	go func() {
		time.Sleep(action.timeout)
		loadErr <- fmt.Errorf("Timeout while loading DOM Content")
	}()
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch v.(type) {
		case *page.EventDomContentEventFired:
			loadErr <- nil
		}
	})
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(action.url),
		})
		if err != nil {
			loadErr <- err
		}
	}()
	return <-loadErr
}

func NavigateAndWaitLoaded(url string, timeout time.Duration) chromedp.Action {
	return &NavigateAndWaitLoadedAction{url, timeout}
}
