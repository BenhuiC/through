package client

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/xtaci/kcp-go"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"through/log"
	"time"
)

const MaxProducer = 20

type ConnectionPool struct {
	ctx     context.Context
	tlsCfg  *tls.Config
	network string
	addr    string
	pool    chan net.Conn
	logger  *log.Logger

	lc          sync.Mutex
	wg          sync.WaitGroup
	producerCnt atomic.Int32
}

func NewConnectionPool(ctx context.Context, size int, network, addr string, tlsCfg *tls.Config) (p *ConnectionPool) {
	p = &ConnectionPool{
		ctx:         ctx,
		pool:        make(chan net.Conn, size),
		network:     network,
		addr:        addr,
		tlsCfg:      tlsCfg,
		logger:      log.NewLogger().With("type", "connectionPool").With("network", network).With("address", addr),
		wg:          sync.WaitGroup{},
		lc:          sync.Mutex{},
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
		p.logger.Debug("get connect timeout, add one producer")
		p.addProducer()
		err = errors.New("timeout")
		return
	case c = <-p.pool:
		// if pool close to null, add producer
		if len(p.pool) <= cap(p.pool)/3 {
			p.logger.Debug("consume too fast, add one producer")
			p.addProducer()
		}
		return
	}
}

func (p *ConnectionPool) addProducer() {
	cnt := p.producerCnt.Load()
	if cnt >= MaxProducer {
		p.logger.Debugf("producer num reach %d, not add", MaxProducer)
		return
	}
	p.lc.Lock()
	defer p.lc.Unlock()
	if cnt == p.producerCnt.Load() {
		p.wg.Add(1)
		go p.producer()
		p.producerCnt.Add(1)
		p.logger.Infof("add connection producer, now is %d", cnt+1)
	}
}

type Producer func(addr string, tlsCfg *tls.Config) (conn net.Conn, err error)

var tcpProducer Producer = func(addr string, tlsCfg *tls.Config) (conn net.Conn, err error) {
	conn, err = tls.Dial("tcp", addr, tlsCfg)
	return
}

var udpProducer Producer = func(addr string, tlsCfg *tls.Config) (conn net.Conn, err error) {
	if conn, err = kcp.Dial(addr); err != nil {
		return
	}
	conn = tls.Client(conn, tlsCfg)
	return
}

func GetProducer(network string) (p Producer) {
	switch network {
	case "tcp":
		return tcpProducer
	case "udp":
		return udpProducer
	}
	return
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
			// sleep 500~1000ms, return if pool still full
			time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)

			// reduce producer
			if v := p.producerCnt.Load(); len(p.pool) == cap(p.pool) && v > 1 {
				p.producerCnt.Add(-1)
				p.logger.Infof("reducer connection producer, now is %d", v-1)
				return
			} else {
				time.Sleep(1 * time.Second)
			}
		}

		start := time.Now()
		// new connection
		prod := GetProducer(p.network)
		if prod == nil {
			p.logger.Errorf("unsupported network %v", p.network)
			return
		}
		c, err := prod(p.addr, p.tlsCfg)
		if err != nil {
			p.logger.Errorf("dial server error:%v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		p.logger.Debugf("produce one connect cost %v", time.Now().Sub(start))

		// return when server is closed
		select {
		case <-p.ctx.Done():
			p.logger.Debug("connection producer stop")
			close(p.pool)
			return
		case p.pool <- c:
			p.logger.Debug("put one connect")
			continue
		}
	}
}

func (p *ConnectionPool) Close() {
	p.logger.Info("close pool")
	for c := range p.pool {
		if c == nil {
			break
		}
		_ = c.Close()
	}
	p.wg.Wait()
}
