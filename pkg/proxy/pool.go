package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
	Alive        bool
	mu           sync.RWMutex
}

// BackendPool manages a collection of backends and routes traffic safely
type BackendPool struct {
	backends []*Backend
	current  uint64 // Using a unit64 to leverage atomic operations safely
}

// NewBackendPool initializes a pool with optimized reverse proxies for each target
func NewBackendPool(targets []string, transport *http.Transport) (*BackendPool, error) {
	pool := &BackendPool{}

	for _, target := range targets {
		targetURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}

		// Configure a dedicated reverse proxy for this specific backend target
		rp := httputil.NewSingleHostReverseProxy(targetURL)
		rp.Transport = transport

		pool.backends = append(pool.backends, &Backend{
			URL:          targetURL,
			ReverseProxy: rp,
			Alive:        true,
		})
	}

	return pool, nil
}

// GetNextPeer uses lock-free atomic round-robin selection to pick the next available backend
func (p *BackendPool) GetNextPeer() *Backend {
	l := len(p.backends)

	for i := 0; i < l; i++ {
		// Increment our counter atomically. This is thread-safe and lock-free.
		// We use the module operator over the length of backends to loop cleanly.
		n := atomic.AddUint64(&p.current, 1)
		idx := (n - 1) % uint64(l)

		target := p.backends[idx]

		// Safely read the status under a read-lock.
		target.mu.RLock()
		isAlive := target.Alive
		target.mu.RUnlock()

		if isAlive {
			return target
		}
	}

	return nil
}

// GetBackends returns all backends in the pool for the health checker to iterate over
func (p *BackendPool) GetBackends() []*Backend {
	return p.backends
}

// SetAlive updates a backend's status safely using a write-lock
func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.Alive = alive
	b.mu.Unlock()
}
