package browser

import (
	"context"
	"strings"
)

type ctxKey string

const avoidProxyKey ctxKey = "avoid_proxy"

func WithAvoidProxy(ctx context.Context, proxy string) context.Context {
	proxy = strings.TrimSpace(proxy)
	if proxy == "" {
		return ctx
	}
	return context.WithValue(ctx, avoidProxyKey, proxy)
}

func avoidProxyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx.Value(avoidProxyKey).(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}
