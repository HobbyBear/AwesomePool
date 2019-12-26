package AwesomePool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Config struct {
	Concurrent  int
	IdleTimeout time.Duration
}

type Pool struct {
	Cfg     *Config
	ctx     context.Context
	cancel  context.CancelFunc
	pending chan func(ctx context.Context)
	workers chan struct{}
	wg      sync.WaitGroup
	closed  bool
	mux     sync.RWMutex
}

func New(ctx context.Context, cfg *Config) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	gr := &Pool{
		Cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		pending: make(chan func(context.Context)),
		workers: make(chan struct{}, cfg.Concurrent),
		wg:      sync.WaitGroup{},
		closed:  false,
		mux:     sync.RWMutex{},
	}
	return gr
}

func (g *Pool) Do(f func(context.Context)) *Pool {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	select {
	case g.pending <- f:
	case g.workers <- struct{}{}:
		g.wg.Add(1)
		go g.loop(f)
	}
	return g
}

func (g *Pool) loop(f func(context.Context)) {
	defer g.wg.Done()
	defer func() { <-g.workers }()
	timer := time.NewTimer(g.Cfg.IdleTimeout)
	defer timer.Stop()
	for {
		g.execute(f)
		timer.Reset(g.Cfg.IdleTimeout)
		select {
		case f = <-g.pending:
			if f == nil {
				return
			}
		case <-timer.C:
			return
		}
	}
}

func (g *Pool) execute(f func(ctx context.Context)) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	f(g.ctx)
}

func (g *Pool) Close(grace bool) {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()
	close(g.pending)
	close(g.workers)
	g.cancel()
	if grace {
		g.wg.Wait()
	}
}
