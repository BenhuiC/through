package client

import (
	"context"
	"crypto/tls"
	"errors"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"through/log"
	"through/proto"
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

func (p *ConnectionPool) getProducer() (pro Producer) {
	switch p.network {
	case "tcp":
		return tcpProducer
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
		prod := p.getProducer()
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

type GrpcConnection struct {
	stream              proto.Through_ForwardClient
	Response            *proto.ForwardResponse
	readOffset          int
	readLock, writeLock sync.Mutex
	responseLock        sync.Mutex
}

func (g *GrpcConnection) Close() error {
	log.Debugf("close grpc connection")
	g.writeLock.Lock()
	defer g.writeLock.Unlock()
	return g.stream.CloseSend()
}

func (g *GrpcConnection) LocalAddr() net.Addr {
	return nil
}

func (g *GrpcConnection) RemoteAddr() net.Addr {
	return nil
}

func (g *GrpcConnection) SetDeadline(t time.Time) error {
	return nil
}

func (g *GrpcConnection) SetReadDeadline(t time.Time) error {
	return nil
}

func (g *GrpcConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

func (g *GrpcConnection) Read(p []byte) (n int, err error) {
	g.readLock.Lock()
	defer g.readLock.Unlock()
	defer func() {
		log.Debugf("read from grpc connection: %v", n)
	}()

	if g.readOffset == 0 {
		if err = g.stream.RecvMsg(g.Response); err != nil {
			log.Errorf("receive strem error: %v", err)
			return 0, err
		}
	}

	data := g.Response.GetData()
	copy(p, data[g.readOffset:])

	if len(data) <= (len(p) + g.readOffset) {
		n := len(data) - g.readOffset
		g.readOffset = 0
		g.Response.Reset()

		return n, err
	}

	g.readOffset += len(p)

	return len(p), nil
}

func (g *GrpcConnection) Write(p []byte) (n int, err error) {
	g.writeLock.Lock()
	defer g.writeLock.Unlock()
	log.Debugf("write data to grpc: %v", len(p))

	g.responseLock.Lock()
	err = g.stream.Send(&proto.ForwardRequest{Data: p})
	g.responseLock.Unlock()
	if err != nil {
		return
	}
	n = len(p)
	return
}
