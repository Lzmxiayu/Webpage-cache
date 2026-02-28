package browser

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type pooledInstance interface {
	Screenshotter
	Close()
}

type PooledScreenshotter struct {
	ch        chan pooledInstance
	instances []pooledInstance
	once      sync.Once
}

func NewPooledChromedpScreenshotter(poolSize, maxTabsPerBrowser int, timeout time.Duration) (*PooledScreenshotter, error) {
	if poolSize <= 0 {
		return nil, fmt.Errorf("pool size must be > 0")
	}
	if maxTabsPerBrowser <= 0 {
		return nil, fmt.Errorf("max tabs per browser must be > 0")
	}

	instances := make([]pooledInstance, 0, poolSize)
	ch := make(chan pooledInstance, poolSize*maxTabsPerBrowser)
	for i := 0; i < poolSize; i++ {
		inst := NewChromedpScreenshotter(timeout)
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

func (p *PooledScreenshotter) Capture(ctx context.Context, url string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case inst := <-p.ch:
		defer func() { p.ch <- inst }()
		return inst.Capture(ctx, url)
	}
}

func (p *PooledScreenshotter) Close() {
	p.once.Do(func() {
		for _, inst := range p.instances {
			inst.Close()
		}
	})
}
