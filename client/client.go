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
	ctx       context.Context
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

	// new http proxy handler
	httpProxy, err := NewHttpProxy(ctx, tlsCfg, cfg.PoolSize, cfg.Servers, cfg.Rulers)
	if err != nil {
		return
	}

	c = &Client{
		ctx:       ctx,
		httpProxy: httpProxy,
		wg:        sync.WaitGroup{},
	}
	return
}

// Start listen and proxy
func (c *Client) Start() (err error) {

	cfg := config.Client
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Info("tcp listener error: %v", err)
		return
	}
	c.listener = listener

	log.Info("client listen at %v", cfg.Addr)
	c.wg.Add(1)
	go c.listenHttp()

	<-c.ctx.Done()
	return
}

func (c *Client) listenHttp() {
	defer c.wg.Done()
	if err := http.Serve(c.listener, c.httpProxy); err != nil {
		log.Error("http server error: %v", err)
	}
}

func (c *Client) Stop() {
	c.httpProxy.Close()
	if c.listener != nil {
		_ = c.listener.Close()
	}
	c.wg.Wait()
}
