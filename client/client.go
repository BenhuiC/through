package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"through/config"
	"through/log"
	"through/util"
)

type Client struct {
	ctx           context.Context
	ruleManager   *RuleManager
	forwardManger *ForwardManger

	httpListener net.Listener
	httpProxy    *HttpProxy

	socksListener net.Listener
	socksProxy    *SocksProxy

	wg sync.WaitGroup
}

func NewClient(ctx context.Context) (c *Client, err error) {
	cfg := config.Client
	var tlsCfg *tls.Config
	tlsCfg, err = util.LoadTlsConfig(cfg.PrivateKey, cfg.CrtFile, "", true)
	if err != nil {
		return
	}

	// new proxy server manager
	forwardManger, err := NewForwardManger(ctx, cfg.Servers, tlsCfg, cfg.PoolSize)
	if err != nil {
		return
	}

	// new host resolver
	resolvers, err := NewResolverManger(ctx, cfg.Resolvers)
	if err != nil {
		return
	}

	// new proxy rule manger
	ruleManger, err := NewRuleManager(resolvers, cfg.Rules)
	if err != nil {
		return
	}

	// new http proxy handler
	httpProxy := NewHttpProxy(ctx, forwardManger, ruleManger)

	// new socks proxy handler
	socksProxy := NewSocksProxy(ctx, forwardManger, ruleManger)

	c = &Client{
		ctx:           ctx,
		httpProxy:     httpProxy,
		socksProxy:    socksProxy,
		wg:            sync.WaitGroup{},
		forwardManger: forwardManger,
		ruleManager:   ruleManger,
	}
	return
}

// Start listen and proxy
func (c *Client) Start() (err error) {

	cfg := config.Client
	// start http listener
	httpLis, err := net.Listen("tcp", cfg.HttpAddr)
	if err != nil {
		log.Infof("tcp http listener error: %v", err)
		return
	}
	c.httpListener = httpLis

	log.Infof("client http listen at %v", cfg.HttpAddr)
	c.wg.Add(1)
	go c.listenHttp()

	// start socks listener
	socksLis, err := net.Listen("tcp", cfg.SocksAddr)
	if err != nil {
		log.Infof("tcp socks listener error: %v", err)
		return
	}
	c.socksListener = socksLis

	log.Infof("client socks listen at %v", cfg.SocksAddr)
	c.wg.Add(1)
	go c.listenSocks()

	<-c.ctx.Done()
	return
}

func (c *Client) listenHttp() {
	defer c.wg.Done()
	if err := http.Serve(c.httpListener, c.httpProxy); err != nil {
		log.Errorf("http server error: %v", err)
	}
}

func (c *Client) listenSocks() {
	defer c.wg.Done()
	for {
		conn, err := c.socksListener.Accept()
		if err != nil {
			log.Errorf("socks listener error: %v", err)
			return
		}
		c.socksProxy.Serve(conn)
	}
}

func (c *Client) Stop() {
	if c.forwardManger != nil {
		log.Info("close forward manger")
		c.forwardManger.Close()
	}
	if c.httpListener != nil {
		log.Info("close http listener")
		_ = c.httpListener.Close()
	}
	if c.socksListener != nil {
		log.Info("close socks listener")
		_ = c.socksListener.Close()
	}
	c.wg.Wait()
}
