package client

import (
	"context"
	"net"
	"through/log"
	"through/proto"
)

type SocksProxy struct {
	forwardManager *ForwardManger
	ruleManager    *RuleManager
}

func NewSocksProxy(ctx context.Context, forwards *ForwardManger, rules *RuleManager) (s *SocksProxy) {
	return &SocksProxy{
		forwardManager: forwards,
		ruleManager:    rules,
	}
}

func (s *SocksProxy) Serve(conn net.Conn) {
	meta, err := s.readMetaFromConn(conn)
	if err != nil {
		log.Errorf("reader meta error: %v", err)
		_ = conn.Close()
		return
	}
	server := s.ruleManager.Get(meta.GetAddress())
	f, ok := s.forwardManager.GetForward(server)
	if !ok {
		log.Infof("host %v math no server", meta.GetAddress())
		_ = conn.Close()
		return
	}

	f.Connect(conn, meta)
}

func (s *SocksProxy) readMetaFromConn(conn net.Conn) (meta *proto.Meta, err error) {
	// todo
	return
}
