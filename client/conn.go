package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"through/log"
	"time"
)

const MaxProducer = 10

type ConnectionPool struct {
	ctx    context.Context
	tlsCfg *tls.Config
	addr   string
	pool   chan net.Conn

	sync.Mutex
	wg          sync.WaitGroup
	producerCnt atomic.Int32
}

func NewConnectionPool(ctx context.Context, tlsCfg *tls.Config, addr string, size int) (p *ConnectionPool) {
	p = &ConnectionPool{
		ctx:         ctx,
		pool:        make(chan net.Conn, size),
		addr:        addr,
		tlsCfg:      tlsCfg,
		wg:          sync.WaitGroup{},
		Mutex:       sync.Mutex{},
		producerCnt: atomic.Int32{},
	}

	p.addProducer()
	return p
}

// Get acquire connection from pool ,return error if timeout
func (p *ConnectionPool) Get(timeout context.Context) (c net.Conn, err error) {
	select {
	case <-p.ctx.Done():
		return
	case <-timeout.Done():
		p.addProducer()
		err = errors.New("timeout")
		return
	case c = <-p.pool:
		return
	}
}

func (p *ConnectionPool) addProducer() {
	cnt := p.producerCnt.Load()
	if cnt >= MaxProducer {
		return
	}
	p.Lock()
	defer p.Unlock()
	if cnt == p.producerCnt.Load() {
		p.wg.Add(1)
		go p.producer()
		p.producerCnt.Add(1)
		log.Infof("add connection producer, now is %d", cnt+1)
	}
}

func (p *ConnectionPool) producer() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}

		// check if pool full
		if len(p.pool) == cap(p.pool) {
			// reduce producer
			if v := p.producerCnt.Load(); v > 1 {
				p.producerCnt.Add(-1)
				log.Infof("reducer connection producer, now is %d", v)
				return
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		// new connection
		c, err := tls.Dial("tcp", p.addr, p.tlsCfg)
		if err != nil {
			log.Errorf("dial server error:%v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// return when server is closed
		select {
		case <-p.ctx.Done():
			return
		default:
			p.pool <- c
		}

	}
}

func (p *ConnectionPool) Close() {
	close(p.pool)
	for c := range p.pool {
		if c == nil {
			break
		}
		_ = c.Close()
	}
	p.wg.Wait()
}
