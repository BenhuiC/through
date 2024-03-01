package server

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/xtaci/kcp-go"
	"go.uber.org/zap"
	"net"
	"sync"
	"through/config"
	"through/log"
	"through/util"
)

type Server struct {
	ctx         context.Context
	tlsCfg      *tls.Config
	tcpListener net.Listener
	udpListener net.Listener
	wg          sync.WaitGroup
}

func NewServer(ctx context.Context) (s *Server, err error) {
	cfg := config.Server
	tlsCfg, err := util.LoadTlsConfig(cfg.PrivateKey, cfg.CrtFile, cfg.CAFile, false)
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
	log.Infof("server start")

	cfg := config.Server
	tcpListener, err := tls.Listen("tcp", cfg.TcpAddr, s.tlsCfg)
	if err != nil {
		log.Infof("tcp listener error: %v", err)
		return
	}
	s.tcpListener = tcpListener

	updListener, err := kcp.Listen(cfg.UdpAddr)
	if err != nil {
		log.Infof("upd listener error: %v", err)
		return
	}
	s.udpListener = updListener

	log.Infof("tcp server listen at %v", cfg.TcpAddr)
	s.wg.Add(1)
	go s.listenTcp()

	log.Infof("upd server listen at %v", cfg.UdpAddr)
	s.wg.Add(1)
	go s.listenUdp()

	<-s.ctx.Done()
	return nil
}

func (s *Server) listenTcp() {
	defer s.wg.Done()
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Errorf("tcp connection accept error: %v", err)
			}
			return
		}

		select {
		case <-s.ctx.Done():
			return
		default:
		}

		log.Infof("accept connection from: %v", conn.RemoteAddr())

		con := NewConnection(s.ctx, conn, log.NewLogger(zap.AddCallerSkip(1)))
		go con.Process()
	}
}

func (s *Server) listenUdp() {
	defer s.wg.Done()
	for {
		conn, err := s.udpListener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Errorf("upd connection accept error: %v", err)
			}
			return
		}

		select {
		case <-s.ctx.Done():
			return
		default:
		}

		log.Infof("accept connection from: %v", conn.RemoteAddr())

		conn = tls.Server(conn, s.tlsCfg)
		con := NewConnection(s.ctx, conn, log.NewLogger(zap.AddCallerSkip(1)))
		go con.Process()
	}
}

func (s *Server) Stop() {
	log.Infof("server stopping")
	if err := s.tcpListener.Close(); err != nil {
		log.Warnf("close server listener error: %v", err)
	}
	s.wg.Wait()
}
