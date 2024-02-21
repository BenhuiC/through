package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
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

	log.Info("server listen %v", cfg.Addr)
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

		go doStuff(ctx, conn)
	}
}

func (s *Server) Stop() {
	log.Info("server stopping")
	if err := s.listener.Close(); err != nil {
		log.Warn("close server listener error: %v", err)
	}
	s.wg.Wait()
}

func doStuff(ctx context.Context, conn net.Conn) {
	//defer conn.Close()

	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		log.Error("reader header error: %v", err)
		return
	}

	dataLen := binary.BigEndian.Uint32(header)

	buf := make([]byte, dataLen)

	if _, err := io.ReadFull(conn, buf); err != nil {
		log.Error("reader data error: %v", err)
		return
	}

	log.Info("receive data: %v", string(buf))

	m := make(map[string]string)
	_ = json.Unmarshal(buf, &m)

	server, err := net.Dial(m["net"], m["addr"])
	if err != nil {
		log.Error("new dial error: %v", err)
		return
	}

	go forward(conn, server)
	go forward(server, conn)
}

func forward(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	_, err := io.Copy(dest, src)
	if err != nil {
		log.Warn("copy connection error: %v", err)
	}

	//time.Sleep(100 * time.Second)
}
