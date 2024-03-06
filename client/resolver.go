package client

import (
	"context"
	"errors"
	"github.com/ncruces/go-dns"
	"golang.org/x/sync/singleflight"
	"net"
	"strings"
	"sync"
	"through/config"
	"through/log"
	"through/util"
	"time"
)

type ResolverManager struct {
	resolvers []*net.Resolver
	cache     map[string]*ResolveCache
	lc        sync.RWMutex
	sf        singleflight.Group
}

type ResolveCache struct {
	ip    net.IP
	addAt time.Time
}

// NewResolverManger manage the resolvers of given config
func NewResolverManger(ctx context.Context, cfg []config.ResolverServer) (r *ResolverManager, err error) {
	r = &ResolverManager{
		resolvers: make([]*net.Resolver, 0, len(cfg)),
		cache:     make(map[string]*ResolveCache),
		lc:        sync.RWMutex{},
		sf:        singleflight.Group{},
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

// Lookup host->ip
func (s *ResolverManager) Lookup(host string) (ip net.IP) {
	start := time.Now()
	defer func() {
		log.Debugf("lookup %v cost %v", host, time.Now().Sub(start))
	}()

	// check cache
	if ip = s.getCache(host); ip != nil {
		log.Debugf("%v get cache %s", host, ip.String())
		return
	}

	// use singleflight
	val, err, _ := s.sf.Do(host, func() (interface{}, error) {
		v := s.doResolver(host)
		return v, nil
	})
	if err != nil {
		log.Errorf("do resolver error: %v", err)
		return
	}

	ip, _ = val.(net.IP)
	if ip != nil {
		s.setCache(host, ip)
	}

	return
}

func (s *ResolverManager) doResolver(host string) (ip net.IP) {
	// do resolve
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resultChan := make(chan net.IP)
	resolve := func(r *net.Resolver) {
		ips, err := r.LookupIP(ctx, "ip", host)
		if err != nil {
			log.Debugf("resolve %v error: %v", host, err)
		}
		var ipRes net.IP
		if len(ips) > 0 {
			ipRes = ips[0]
		} else {
			return
		}
		select {
		case resultChan <- ipRes:
			return
		case <-ctx.Done():
			return
		}
	}

	for i := range s.resolvers {
		go resolve(s.resolvers[i])
	}

	// wait until a success result or just timeout
	select {
	case ip = <-resultChan:
		// cancel to stop all goroutine
		cancel()
		return
	case <-ctx.Done():
		return
	}
}

// Country get the country where the host located at, return IsoCode
func (s *ResolverManager) Country(host string) (c string) {
	ipAddr := s.Lookup(host)
	c = util.Country(ipAddr)
	return
}

func (s *ResolverManager) getCache(host string) (ip net.IP) {
	s.lc.RLock()
	defer s.lc.RUnlock()
	if i, ok := s.cache[host]; ok {
		ip = i.ip
	}
	return
}

// setCache set host->ip in cache map, expire at 30 seconds
func (s *ResolverManager) setCache(host string, ip net.IP) {
	s.lc.Lock()
	defer s.lc.Unlock()
	s.cache[host] = &ResolveCache{
		ip:    ip,
		addAt: time.Now(),
	}
}

func (s *ResolverManager) cleanUp(ctx context.Context) {
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case now := <-ticker:
			s.lc.Lock()
			for k, v := range s.cache {
				if now.Sub(v.addAt) > 30*time.Second {
					delete(s.cache, k)
					log.Debugf("delete cache %v", k)
				}
			}
			s.lc.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// NewLocalResolver default host resolver with system config
func NewLocalResolver() (n *net.Resolver) {
	n = &net.Resolver{}
	return
}

// NewDNSResolver new dns resolver with give server
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

// NewDoTResolver new resolver with give server on tls
func NewDoTResolver(server string) (n *net.Resolver, err error) {
	n, err = dns.NewDoTResolver(server)
	return
}
