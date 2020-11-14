package actions

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/markoczy/crawler/types"
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
	chErr := types.NewErrorSwitchChannel()
	go func() {
		time.Sleep(action.timeout)
		chErr.Send(fmt.Errorf("Timeout while loading DOM Content"))
	}()
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch v.(type) {
		case *page.EventDomContentEventFired:
			chErr.Send(nil)
		}
	})
	go func() {
		err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(action.url),
		})
		if err != nil {
			chErr.Send(err)
		}
	}()
	err := chErr.Receive()
	return err
}

func NavigateAndWaitLoaded(url string, timeout time.Duration) chromedp.Action {
	return &NavigateAndWaitLoadedAction{url, timeout}
}
