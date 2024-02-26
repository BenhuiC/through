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
	cache     map[string]net.IP // todo
	lc        sync.RWMutex
}

func NewResolverManger(cfg []config.ResolverServer) (r *ResolverManager, err error) {
	r = &ResolverManager{
		resolvers: make([]*net.Resolver, 0, len(cfg)),
		cache:     make(map[string]net.IP),
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

	return
}

func (s *ResolverManager) Lookup(host string) (ip net.IP) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	result := make(chan net.IP)
	resolve := func(r *net.Resolver) {
		defer wg.Done()
		ips, err := r.LookupIP(ctx, "ip", host)
		if err != nil || len(ips) == 0 {
			return
		}
		select {
		case result <- ips[0]:
			return
		case <-ctx.Done():
			return
		}
	}

	// FIXME 超时或error情况会阻塞
	for i := range s.resolvers {
		wg.Add(1)
		go resolve(s.resolvers[i])
	}

	ip = <-result
	cancel()
	close(result)

	wg.Wait()

	return
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
