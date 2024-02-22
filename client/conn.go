package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"through/log"
	"time"
)

type ConnectionPool struct {
	ctx    context.Context
	tlsCfg *tls.Config
	addr   string
	pool   chan net.Conn
	wg     sync.WaitGroup
}

func NewConnectionPool(ctx context.Context, tlsCfg *tls.Config, addr string, size int) (p *ConnectionPool) {
	p = &ConnectionPool{
		ctx:    ctx,
		pool:   make(chan net.Conn, size),
		addr:   addr,
		tlsCfg: tlsCfg,
		wg:     sync.WaitGroup{},
	}
	p.wg.Add(1)
	go p.newConnection()
	return p
}

func (p *ConnectionPool) Get(timeout time.Duration) (c net.Conn, err error) {
	select {
	case <-p.ctx.Done():
		return
	case <-time.After(timeout):
		err = errors.New("timeout")
		return
	case c = <-p.pool:
		return
	}
}

func (p *ConnectionPool) newConnection() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			c, err := tls.Dial("tcp", p.addr, p.tlsCfg)
			if err != nil {
				log.Error("dial server error:%v", err)
				time.Sleep(10 * time.Second)
				continue
			}
			p.pool <- c
		}
	}
}

func (p *ConnectionPool) Stop() {
	close(p.pool)
	p.wg.Wait()
}
