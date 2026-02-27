package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type ChromedpScreenshotter struct {
	browserCtx context.Context
	cancel     context.CancelFunc
	timeout    time.Duration
}

func NewChromedpScreenshotter(timeout time.Duration) *ChromedpScreenshotter {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	return &ChromedpScreenshotter{
		browserCtx: browserCtx,
		cancel: func() {
			browserCancel()
			allocCancel()
		},
		timeout: timeout,
	}
}

func (s *ChromedpScreenshotter) Capture(ctx context.Context, url string) ([]byte, error) {
	tabCtx, tabCancel := chromedp.NewContext(s.browserCtx)
	defer tabCancel()

	if ctx != nil {
		var cancel context.CancelFunc
		tabCtx, cancel = context.WithCancel(tabCtx)
		defer cancel()
		go func() {
			select {
			case <-ctx.Done():
				cancel()
			case <-tabCtx.Done():
			}
		}()
	}

	if s.timeout > 0 {
		var cancel context.CancelFunc
		tabCtx, cancel = context.WithTimeout(tabCtx, s.timeout)
		defer cancel()
	}

	var png []byte
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.FullScreenshot(&png, 90),
	)
	if err != nil {
		return nil, err
	}

	return png, nil
}

func (s *ChromedpScreenshotter) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}
