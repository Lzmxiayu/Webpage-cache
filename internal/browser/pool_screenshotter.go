package browser

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type pooledInstance interface {
	Screenshotter
	Close()
	Proxy() string
}

type PooledScreenshotter struct {
	ch        chan pooledInstance
	instances []pooledInstance
	once      sync.Once
}

func NewPooledChromedpScreenshotter(poolSize, maxTabsPerBrowser int, timeout time.Duration, proxies []string) (*PooledScreenshotter, error) {
	if poolSize <= 0 {
		return nil, fmt.Errorf("pool size must be > 0")
	}
	if maxTabsPerBrowser <= 0 {
		return nil, fmt.Errorf("max tabs per browser must be > 0")
	}

	instances := make([]pooledInstance, 0, poolSize)
	ch := make(chan pooledInstance, poolSize*maxTabsPerBrowser)
	proxyList := normalizedProxies(proxies)
	for i := 0; i < poolSize; i++ {
		proxy := pickProxy(proxyList, i)
		inst := NewChromedpScreenshotter(timeout, proxy)
		instances = append(instances, inst)
		for t := 0; t < maxTabsPerBrowser; t++ {
			ch <- inst
		}
	}

	return &PooledScreenshotter{
		ch:        ch,
		instances: instances,
	}, nil
}

func (p *PooledScreenshotter) Capture(ctx context.Context, url string) ([]byte, string, error) {
	avoidProxy := avoidProxyFromContext(ctx)

	select {
	case <-ctx.Done():
		return nil, "", ctx.Err()
	case inst := <-p.ch:
		chosen := p.pickInstanceAvoiding(inst, avoidProxy)
		defer func() { p.ch <- chosen }()
		return chosen.Capture(ctx, url)
	}
}

func (p *PooledScreenshotter) Close() {
	p.once.Do(func() {
		for _, inst := range p.instances {
			inst.Close()
		}
	})
}

func normalizedProxies(proxies []string) []string {
	out := make([]string, 0, len(proxies))
	for _, p := range proxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func pickProxy(proxies []string, i int) string {
	if len(proxies) == 0 {
		return ""
	}
	return proxies[i%len(proxies)]
}

func (p *PooledScreenshotter) pickInstanceAvoiding(first pooledInstance, avoidProxy string) pooledInstance {
	if avoidProxy == "" || first.Proxy() != avoidProxy {
		return first
	}

	available := len(p.ch)
	held := make([]pooledInstance, 0, available+1)
	held = append(held, first)

	var chosen pooledInstance
	for i := 0; i < available; i++ {
		candidate := <-p.ch
		if chosen == nil && candidate.Proxy() != avoidProxy {
			chosen = candidate
			continue
		}
		held = append(held, candidate)
	}

	for _, inst := range held {
		p.ch <- inst
	}
	if chosen != nil {
		return chosen
	}

	return <-p.ch
}
