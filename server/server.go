package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"io"
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

		log.Info("accept connection from: %v", conn.RemoteAddr())

		go doStuff(ctx, conn)
	}
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

func Stop() {

}
