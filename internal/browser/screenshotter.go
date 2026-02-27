package browser

import "context"

type Screenshotter interface {
	Capture(ctx context.Context, url string) ([]byte, error)
}

