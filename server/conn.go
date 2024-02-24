package server

import (
	"context"
	"net"
	"through/log"
	"through/proto"
	"through/util"
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
		return
	}

	// forward
	util.CopyLoopWait(remote, c.conn)
}
