package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"through/pkg/log"
	"through/pkg/proto"
	"through/util"
)

type TcpServer struct {
	wg       sync.WaitGroup
	listener net.Listener
}

func (t *TcpServer) Start(ctx context.Context, addr string, tlsCfg *tls.Config) (err error) {
	t.listener, err = tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		log.Infof("tcp listener error: %v", err)
		return
	}
	log.Infof("tcp server listen at %v", addr)

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			conn, err := t.listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					log.Errorf("tcp connection accept error: %v", err)
				}
				return
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			log.Infof("accept connection from: %v", conn.RemoteAddr())

			go t.Forward(ctx, conn)
		}
	}()

	return
}

func (t *TcpServer) Forward(ctx context.Context, conn net.Conn) {
	//  read data
	meta, err := proto.ReadMeta(conn)
	if err != nil {
		log.Errorf("read meta data error: %v", err)
		return
	}

	// dial connection
	remote, err := net.Dial(meta.GetNet(), meta.GetAddress())
	if err != nil {
		log.Errorf("dial to %v:%v error:%v", meta.GetNet(), meta.GetAddress(), err)
		_ = conn.Close()
		return
	}
	log.Infof("dial to %v,%v", meta.GetNet(), meta.Address)

	// forward
	util.CopyLoopWait(remote, conn)
}

func (t *TcpServer) Stop() {
	if t == nil || t.listener == nil {
		return
	}
	_ = t.listener.Close()
	t.wg.Wait()
}
