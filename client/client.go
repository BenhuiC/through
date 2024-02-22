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
	ctx            context.Context
	listener       net.Listener
	connectionPool *ConnectionPool
	wg             sync.WaitGroup
}

func NewClient(ctx context.Context) (c *Client, err error) {
	cfg := config.Client
	var tlsCfg *tls.Config
	tlsCfg, err = util.LoadTlsConfig(cfg.PrivateKey, cfg.CrtFile, "", true)
	if err != nil {
		return
	}

	c = &Client{
		ctx:            ctx,
		connectionPool: NewConnectionPool(ctx, tlsCfg, cfg.Server, 10),
		wg:             sync.WaitGroup{},
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
	go c.listenHttp()

	<-c.ctx.Done()
	return
}

func (c *Client) listenHttp() {
	defer c.wg.Done()
	handler := &HttpHandler{
		connPool: c.connectionPool,
	}
	if err := http.Serve(c.listener, handler); err != nil {
		log.Error("http server error: %v", err)
	}
}

func (c *Client) Stop() {
	c.connectionPool.Close()
	if c.listener != nil {
		_ = c.listener.Close()
	}
	c.wg.Wait()
}
