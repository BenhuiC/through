package server

import (
	"context"
	"crypto/tls"
	"through/config"
	"through/pkg/log"
	"through/pkg/proto"
	"through/util"
)

type Server struct {
	proto.UnimplementedThroughServer
	tlsCfg     *tls.Config
	tcpServer  ForwardServer
	grpcServer ForwardServer
}

func NewServer() (s *Server, err error) {
	cfg := config.Server
	tlsCfg, err := util.LoadTlsConfig(cfg.PrivateKey, cfg.CrtFile, cfg.CAFile, false)
	if err != nil {
		return
	}

	s = &Server{
		tlsCfg:     tlsCfg,
		tcpServer:  &TcpServer{},
		grpcServer: &GrpcServer{},
	}

	return
}

func (s *Server) Start(ctx context.Context) (err error) {
	log.Infof("server start")

	cfg := config.Server
	// raw tcp
	if cfg.TcpAddr != "" {
		if err = s.tcpServer.Start(ctx, cfg.TcpAddr, s.tlsCfg); err != nil {
			log.Errorf("start tcp server error: %v", err)
			return
		}
	}

	// grpc
	if cfg.GrpcAddr != "" {
		if err = s.grpcServer.Start(ctx, cfg.GrpcAddr, s.tlsCfg); err != nil {
			log.Errorf("start grpc server error: %v", err)
			return
		}
	}

	<-ctx.Done()
	return nil
}

func (s *Server) Stop() {
	log.Infof("server stopping")
	s.tcpServer.Stop()
	s.grpcServer.Stop()
}

type ForwardServer interface {
	Start(ctx context.Context, addr string, tlsCfg *tls.Config) (err error)
	Stop()
}
