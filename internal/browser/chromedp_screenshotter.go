package browser

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type ChromedpScreenshotter struct {
	browserCtx context.Context
	cancel     context.CancelFunc
	timeout    time.Duration
	renderWait time.Duration
	proxyURL   string
}

func NewChromedpScreenshotter(timeout, renderWait time.Duration, proxyURL string) *ChromedpScreenshotter {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)
	if strings.TrimSpace(proxyURL) != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(strings.TrimSpace(proxyURL)))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	return &ChromedpScreenshotter{
		browserCtx: browserCtx,
		cancel: func() {
			browserCancel()
			allocCancel()
		},
		timeout:    timeout,
		renderWait: renderWait,
		proxyURL:   strings.TrimSpace(proxyURL),
	}
}

func (s *ChromedpScreenshotter) Capture(ctx context.Context, url string) ([]byte, string, error) {
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
		chromedp.Poll(`document.readyState === "complete"`, nil, chromedp.WithPollingInterval(100*time.Millisecond)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if s.renderWait <= 0 {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(s.renderWait):
				return nil
			}
		}),
		chromedp.FullScreenshot(&png, 90),
	)
	if err != nil {
		return nil, s.proxyURL, err
	}

	return png, s.proxyURL, nil
}

func (s *ChromedpScreenshotter) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ChromedpScreenshotter) Proxy() string {
	return s.proxyURL
}
