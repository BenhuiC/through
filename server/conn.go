package server

import (
	"context"
	"io"
	"net"
	"sync"
	"through/log"
	"through/proto"
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
		log.Error("read meta data error: %v", err)
		return
	}

	// dial connection
	remote, err := net.Dial(meta.GetNet(), meta.GetAddress())
	if err != nil {
		log.Error("dial to %v:%v error:%v", meta.GetNet(), meta.GetAddress(), err)
		return
	}

	// forward
	CopyLoopWait(remote, c.conn)
}

func CopyLoopWait(c1 net.Conn, c2 net.Conn) {
	var wg sync.WaitGroup
	cp := func(dst, src net.Conn) {
		defer wg.Done()
		_, err := io.Copy(dst, src)
		_ = dst.Close()
		if err != nil {
			_ = src.Close()
		}
	}
	wg.Add(2)
	go cp(c1, c2)
	go cp(c2, c1)
	wg.Wait()
}
