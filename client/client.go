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
	ctx      context.Context
	listener net.Listener
	tlsCfg   *tls.Config
	wg       sync.WaitGroup
}

func NewClient(ctx context.Context) (c *Client, err error) {
	cfg := config.Client
	var tlsCfg *tls.Config
	tlsCfg, err = util.LoadTlsConfig(cfg.PrivateKey, cfg.CrtFile, "", true)
	if err != nil {
		return
	}

	c = &Client{
		ctx:    ctx,
		tlsCfg: tlsCfg,
		wg:     sync.WaitGroup{},
	}
	return
}

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

	return
}

func (c *Client) listenHttp() {
	defer c.wg.Done()
	http.Serve(c.listener, func() {})
}

func (c *Client) Stop() {
	if c.listener != nil {
		_ = c.listener.Close()
	}
	c.wg.Wait()
}
