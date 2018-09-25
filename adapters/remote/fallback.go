package adapters

import (
	"errors"
	"sync"
	"time"

	C "github.com/Dreamacro/clash/constant"
)

type proxy struct {
	RawProxy C.Proxy
	Valid    bool
}

type Fallback struct {
	name    string
	proxies []*proxy
	rawURL  string
	delay   time.Duration
	done    chan struct{}
}

func (f *Fallback) Name() string {
	return f.name
}

func (f *Fallback) Type() C.AdapterType {
	return C.Fallback
}

func (f *Fallback) Now() string {
	_, proxy := f.findNextValidProxy(0)
	if proxy != nil {
		return proxy.RawProxy.Name()
	}
	return f.proxies[0].RawProxy.Name()
}

func (f *Fallback) Generator(addr *C.Addr) (adapter C.ProxyAdapter, err error) {
	idx := 0
	var proxy *proxy
	for {
		idx, proxy = f.findNextValidProxy(idx)
		if proxy == nil {
			break
		}
		adapter, err = proxy.RawProxy.Generator(addr)
		if err != nil {
			proxy.Valid = false
			idx++
			continue
		}
		return
	}
	return nil, errors.New("There are no valid proxy")
}

func (f *Fallback) Close() {
	f.done <- struct{}{}
}

func (f *Fallback) loop() {
	tick := time.NewTicker(f.delay)
	go f.validTest()
Loop:
	for {
		select {
		case <-tick.C:
			go f.validTest()
		case <-f.done:
			break Loop
		}
	}
}

func (f *Fallback) findNextValidProxy(start int) (int, *proxy) {
	for i := start; i < len(f.proxies); i++ {
		if f.proxies[i].Valid {
			return i, f.proxies[i]
		}
	}
	return -1, nil
}

func (f *Fallback) validTest() {
	wg := sync.WaitGroup{}
	wg.Add(len(f.proxies))

	for _, p := range f.proxies {
		go func(p *proxy) {
			_, err := DelayTest(p.RawProxy, f.rawURL)
			p.Valid = err == nil
			wg.Done()
		}(p)
	}

	wg.Wait()
}

func NewFallback(name string, proxies []C.Proxy, rawURL string, delay time.Duration) (*Fallback, error) {
	_, err := urlToAddr(rawURL)
	if err != nil {
		return nil, err
	}

	if len(proxies) < 1 {
		return nil, errors.New("The number of proxies cannot be 0")
	}

	warpperProxies := make([]*proxy, len(proxies))
	for idx := range proxies {
		warpperProxies[idx] = &proxy{
			RawProxy: proxies[idx],
			Valid:    true,
		}
	}

	Fallback := &Fallback{
		name:    name,
		proxies: warpperProxies,
		rawURL:  rawURL,
		delay:   delay,
		done:    make(chan struct{}),
	}
	go Fallback.loop()
	return Fallback, nil
}