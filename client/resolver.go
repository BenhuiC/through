package client

import (
	"context"
	"errors"
	"github.com/ncruces/go-dns"
	"net"
	"strings"
	"sync"
	"through/config"
	"time"
)

type ResolverManager struct {
	resolvers []*net.Resolver
	cache     map[string]*ResolveCache
	lc        sync.RWMutex
}

type ResolveCache struct {
	ip    net.IP
	addAt time.Time
}

func NewResolverManger(ctx context.Context, cfg []config.ResolverServer) (r *ResolverManager, err error) {
	r = &ResolverManager{
		resolvers: make([]*net.Resolver, 0, len(cfg)),
		cache:     make(map[string]*ResolveCache),
		lc:        sync.RWMutex{},
	}

	for _, c := range cfg {
		if c.DoT != "" {
			resolver, err := NewDoTResolver(c.DoT)
			if err != nil {
				return nil, err
			}
			r.resolvers = append(r.resolvers, resolver)
		} else if c.DNS != "" {
			r.resolvers = append(r.resolvers, NewDNSResolver(c.DNS))
		} else {
			err = errors.New("resolver config error")
			return
		}
	}

	if len(r.resolvers) == 0 {
		r.resolvers = append(r.resolvers, NewLocalResolver())
	}

	go r.cleanUp(ctx)
	return
}

func (s *ResolverManager) Lookup(host string) (ip net.IP) {
	// check cache
	if ip = s.getCache(host); ip != nil {
		return
	}

	// do resolve
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	type resolveResult struct {
		net.IP
		error
	}
	resultChan := make(chan resolveResult)
	resolve := func(r *net.Resolver) {
		defer wg.Done()
		ips, err := r.LookupIP(ctx, "ip", host)
		var ipRes net.IP
		if len(ips) > 0 {
			ipRes = ips[0]
		}
		select {
		case resultChan <- resolveResult{IP: ipRes, error: err}:
			return
		case <-ctx.Done():
			return
		}
	}

	for i := range s.resolvers {
		wg.Add(1)
		go resolve(s.resolvers[i])
	}

	for {
		// wait until a success result or just timeout
		select {
		case res := <-resultChan:
			if res.error != nil {
				continue
			}
			ip = res.IP
			s.setCache(host, ip)
			// cancel to stop all goroutine
			cancel()
			wg.Wait()
			close(resultChan)
			return
		case <-ctx.Done():
			return
		}
	}

}

func (s *ResolverManager) getCache(host string) (ip net.IP) {
	s.lc.RLock()
	defer s.lc.RUnlock()
	if i, ok := s.cache[host]; ok {
		ip = i.ip
	}
	return
}

func (s *ResolverManager) setCache(host string, ip net.IP) {
	s.lc.Lock()
	defer s.lc.Unlock()
	s.cache[host] = &ResolveCache{
		ip:    ip,
		addAt: time.Now(),
	}
}

func (s *ResolverManager) cleanUp(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			// FIXME ticker not work ?
			now := time.Now()
			s.lc.Lock()
			for k, v := range s.cache {
				if now.Sub(v.addAt) > 30*time.Second {
					delete(s.cache, k)
				}
			}
			s.lc.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func NewLocalResolver() (n *net.Resolver) {
	n = &net.Resolver{}
	return
}

func NewDNSResolver(server string) (n *net.Resolver) {
	if !strings.Contains(server, ":") {
		server = server + ":53"
	}

	var d net.Dialer
	n = &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return d.DialContext(ctx, network, server)
		},
	}

	return
}

func NewDoTResolver(server string) (n *net.Resolver, err error) {
	n, err = dns.NewDoTResolver(server)
	return
}
