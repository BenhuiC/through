package server

import (
	"context"
	"crypto/tls"
	"net"
	"through/config"
	"through/log"
)

func Start(ctx context.Context) (err error) {
	log.Info("server start")

	cfg := config.Server
	tlsCfg, err := loadTlsConfig(cfg.PrivateKey, cfg.CrtFile, cfg.CAFile)
	listener, err := tls.Listen("tcp", cfg.Addr, tlsCfg)
	if err != nil {
		log.Info("tcp listener error: %v", err)
		return
	}

	log.Info("server listen %v", cfg.Addr)
	go listen(ctx, listener)

	return nil
}

func listen(ctx context.Context, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("tcp connection accept error: %v", err)
			return
		}

		log.Debug("accept connection from:%v", conn.RemoteAddr())

		go doStuff(ctx, conn)
	}
}

func doStuff(ctx context.Context, conn net.Conn) {
	buf := make([]byte, 0, 100)
	for {
		if _, err := conn.Read(buf); err != nil {
			log.Error("reader data error: %v", err)
			return
		}
		log.Info("receive data: %v", string(buf))
		_, _ = conn.Write([]byte("hello"))
	}
}

func Stop() {

}
