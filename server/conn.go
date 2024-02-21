package server

import (
	"context"
	"net"
	"through/log"
)

type Connection struct {
	conn net.Conn
	ctx  context.Context
	log.Logger
}

func NewConnection(ctx context.Context, conn net.Conn, logger log.Logger) *Connection {
	return &Connection{ctx: ctx, conn: conn, Logger: logger}
}

func (c *Connection) Process() {
	// todo reader data

	// todo dial connection

	// todo forward
}
