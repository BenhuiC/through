package server

import (
	"context"
	"net"
	"sync"
	"through/pkg/log"
	proto "through/pkg/proto"
	"through/util"
	"time"
)

type Connection struct {
	conn net.Conn
	ctx  context.Context
	*log.Logger
}

func NewConnection(ctx context.Context, conn net.Conn, logger *log.Logger) (c *Connection) {
	return &Connection{ctx: ctx, conn: conn, Logger: logger}
}

func (c *Connection) Process() {
	//  read data
	meta, err := proto.ReadMeta(c.conn)
	if err != nil {
		log.Errorf("read meta data error: %v", err)
		return
	}

	// dial connection
	remote, err := net.Dial(meta.GetNet(), meta.GetAddress())
	if err != nil {
		log.Errorf("dial to %v:%v error:%v", meta.GetNet(), meta.GetAddress(), err)
		_ = c.conn.Close()
		return
	}
	log.Infof("dial to %v,%v", meta.GetNet(), meta.Address)

	// forward
	util.CopyLoopWait(remote, c.conn)
}

type GrpcConnection struct {
	stream              proto.Through_ForwardServer
	Response            *proto.ForwardRequest
	readOffset          int
	readLock, writeLock sync.Mutex
	responseLock        sync.Mutex
}

func (g *GrpcConnection) Close() error {
	log.Debugf("close grpc connection")
	return nil
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
	defer func() {
		log.Debugf("read from grpc connection: %v", n)
	}()
	g.readLock.Lock()
	defer g.readLock.Unlock()

	if g.readOffset == 0 {
		if err = g.stream.RecvMsg(g.Response); err != nil {
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
	err = g.stream.Send(&proto.ForwardResponse{Data: p})
	g.responseLock.Unlock()
	if err != nil {
		return
	}
	n = len(p)
	return
}
