package server

import (
	"context"
	"crypto/tls"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"net"
	"sync"
	"through/pkg/log"
	"through/pkg/proto"
	"through/util"
	"time"
)

type GrpcServer struct {
	proto.UnimplementedThroughServer
	wg       sync.WaitGroup
	listener net.Listener
}

func (s *GrpcServer) Start(ctx context.Context, addr string, tlsCfg *tls.Config) (err error) {
	s.listener, err = tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		log.Infof("grpc listener error: %v", err)
		return
	}
	log.Infof("grpc server listen at %v", addr)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		var kaep = keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true, // Allow pings even when there are no active streams
		}
		var kasp = keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 3 * time.Second,
		}
		server := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
		proto.RegisterThroughServer(server, s)
		err := server.Serve(s.listener)
		if err != nil {
			log.Errorf("grcp connection serve error: %v", err)
		}
	}()
	return
}

func (s *GrpcServer) Stop() {
	if s == nil || s.listener == nil {
		return
	}
	_ = s.listener.Close()
	s.wg.Wait()
}

func (s *GrpcServer) Forward(stream proto.Through_ForwardServer) (err error) {
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
