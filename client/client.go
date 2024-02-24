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

	listener  net.Listener
	httpProxy *HttpProxy

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

	// new proxy rule manger
	ruleManger, err := NewRuleManager(cfg.Rules)
	if err != nil {
		return
	}

	// new http proxy handler
	httpProxy := NewHttpProxy(ctx, forwardManger, ruleManger)

	c = &Client{
		ctx:           ctx,
		httpProxy:     httpProxy,
		wg:            sync.WaitGroup{},
		forwardManger: forwardManger,
		ruleManager:   ruleManger,
	}
	return
}

// Start listen and proxy
func (c *Client) Start() (err error) {

	cfg := config.Client
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Infof("tcp listener error: %v", err)
		return
	}
	c.listener = listener

	log.Infof("client listen at %v", cfg.Addr)
	c.wg.Add(1)
	go c.listenHttp()

	<-c.ctx.Done()
	return
}

func (c *Client) listenHttp() {
	defer c.wg.Done()
	if err := http.Serve(c.listener, c.httpProxy); err != nil {
		log.Errorf("http server error: %v", err)
	}
}

func (c *Client) Stop() {
	if c.forwardManger != nil {
		c.forwardManger.Close()
	}
	if c.listener != nil {
		_ = c.listener.Close()
	}
	c.wg.Wait()
}
