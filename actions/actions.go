package actions

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"golang.org/x/exp/errors/fmt"
)

func recoverChannelClosed() {
	recover()
}

type NavigateAndWaitLoadedAction struct {
	url     string
	timeout time.Duration
}

func (action *NavigateAndWaitLoadedAction) Do(ctx context.Context) error {
	loadErr := make(chan error, 10)
	go func() {
		defer recoverChannelClosed()
		time.Sleep(action.timeout)
		loadErr <- fmt.Errorf("Timeout while loading DOM Content")
	}()
	chromedp.ListenTarget(ctx, func(v interface{}) {
		defer recoverChannelClosed()
		switch v.(type) {
		case *page.EventDomContentEventFired:
			loadErr <- nil
		}
	})
	go func() {
		defer recoverChannelClosed()
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(action.url),
		})
		if err != nil {
			loadErr <- err
		}
	}()
	err := <-loadErr
	close(loadErr)
	return err
}

func NavigateAndWaitLoaded(url string, timeout time.Duration) chromedp.Action {
	return &NavigateAndWaitLoadedAction{url, timeout}
}
