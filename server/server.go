package server

import (
	"context"
	"crypto/tls"
	"errors"
	"go.uber.org/zap"
	"net"
	"sync"
	"through/config"
	"through/log"
)

type Server struct {
	ctx      context.Context
	tlsCfg   *tls.Config
	listener net.Listener
	wg       sync.WaitGroup
}

func NewServer(ctx context.Context) (s *Server, err error) {
	cfg := config.Server
	tlsCfg, err := loadTlsConfig(cfg.PrivateKey, cfg.CrtFile, cfg.CAFile)
	if err != nil {
		return
	}

	s = &Server{
		ctx:    ctx,
		tlsCfg: tlsCfg,
		wg:     sync.WaitGroup{},
	}

	return
}

func (s *Server) Start() (err error) {
	log.Info("server start")

	cfg := config.Server
	listener, err := tls.Listen("tcp", cfg.Addr, s.tlsCfg)
	if err != nil {
		log.Info("tcp listener error: %v", err)
		return
	}
	s.listener = listener

	log.Info("server listen at %v", cfg.Addr)
	s.wg.Add(1)
	go s.listen(s.ctx, listener)

	return nil
}

func (s *Server) listen(ctx context.Context, listener net.Listener) {
	defer s.wg.Done()
	for {
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Error("tcp connection accept error: %v", err)
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Info("accept connection from: %v", conn.RemoteAddr())

		con := NewConnection(ctx, conn, log.NewLogger(zap.AddCallerSkip(1)))
		go con.Process()
	}
}

func (s *Server) Stop() {
	log.Info("server stopping")
	if err := s.listener.Close(); err != nil {
		log.Warn("close server listener error: %v", err)
	}
	s.wg.Wait()
}
