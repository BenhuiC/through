package server

import (
	"context"
	"crypto/tls"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"net"
	"sync"
	"through/config"
	"through/pkg/log"
	"through/pkg/proto"
	"through/util"
	"time"
)

type Server struct {
	proto.UnimplementedThroughServer
	ctx          context.Context
	tlsCfg       *tls.Config
	tcpListener  net.Listener
	grpcListener net.Listener
	wg           sync.WaitGroup
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
	// raw tcp
	if cfg.TcpAddr != "" {
		s.tcpListener, err = tls.Listen("tcp", cfg.TcpAddr, s.tlsCfg)
		if err != nil {
			log.Infof("tcp listener error: %v", err)
			return
		}
		log.Infof("tcp server listen at %v", cfg.TcpAddr)
		s.wg.Add(1)
		go s.listenTcp()
	}

	// grpc
	if cfg.GrpcAddr != "" {
		s.grpcListener, err = tls.Listen("tcp", cfg.GrpcAddr, s.tlsCfg)
		if err != nil {
			log.Infof("grpc listener error: %v", err)
			return
		}
		log.Infof("grpc server listen at %v", cfg.GrpcAddr)
		s.wg.Add(1)
		go s.listenGrpc()
	}

	<-s.ctx.Done()
	return nil
}

func (s *Server) Forward(stream proto.Through_ForwardServer) (err error) {
	var remote net.Conn

	if remote == nil {
		var req *proto.ForwardRequest
		req, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		} else if err != nil {
			log.Errorf("receive from stream error: %v", err)
			return
		}
		log.Infof("start connection: %v", req.GetMeta())
		if req.GetMeta() == nil {
			err = errors.New("first request must contain meta")
			return
		}
		meta := req.GetMeta()
		remote, err = net.Dial(meta.GetNet(), meta.GetAddress())
		if err != nil {
			log.Errorf("dial to %v:%v error:%v", meta.GetNet(), meta.GetAddress(), err)
			return err
		}
	}

	util.CopyLoopWait(remote, &GrpcConnection{stream: stream, Response: &proto.ForwardRequest{}})
	log.Info("close connection")
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

func (s *Server) listenGrpc() {
	defer s.wg.Done()
	var kaep = keepalive.EnforcementPolicy{
		PermitWithoutStream: true, // Allow pings even when there are no active streams
	}
	var kasp = keepalive.ServerParameters{
		Time:    30 * time.Second,
		Timeout: 3 * time.Second,
	}
	server := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	proto.RegisterThroughServer(server, s)
	err := server.Serve(s.grpcListener)
	if err != nil {
		log.Errorf("grcp connection serve error: %v", err)
	}
}

func (s *Server) Stop() {
	log.Infof("server stopping")
	if s.tcpListener != nil {
		if err := s.tcpListener.Close(); err != nil {
			log.Warnf("close tcp listener error: %v", err)
		}
	}
	if s.grpcListener != nil {
		if err := s.grpcListener.Close(); err != nil {
			log.Warnf("close grcp listener error: %v", err)
		}
	}
	s.wg.Wait()
}
