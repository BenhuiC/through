package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"through/log"
	"through/proto"
)

const (
	Socks5Version = 0x05

	SocksNoAuthentication    = 0x00
	SocksNoAcceptableMethods = 0xFF

	SocksIPv4Host   = 0x01
	SocksIPv6Host   = 0x04
	SocksDomainHost = 0x03

	SocksCmdConnect      = 0x01
	SocksCmdBind         = 0x02
	SocksCmdUDPAssociate = 0x03
)

var (
	UnSupportVersion = errors.New("unsupported socks version")
	UnSupportCommand = errors.New("unsupported command")
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
	if err = s.auth(conn); err != nil {
		return
	}
	return s.connect(conn)
}

func (s *SocksProxy) auth(conn net.Conn) (err error) {
	buf := make([]byte, 256)

	// 读取 VER 和 NMETHODS
	_, err = io.ReadFull(conn, buf[:2])
	if err != nil {
		return
	}

	ver, nMethods := buf[0], buf[1]
	if ver != Socks5Version {
		err = UnSupportVersion
		return
	}

	// 读取 METHODS 列表
	_, err = io.ReadFull(conn, buf[:nMethods])
	if err != nil {
		return
	}

	// write response
	_, err = conn.Write([]byte{Socks5Version, SocksNoAuthentication})

	// todo support auth

	return
}

func (s *SocksProxy) connect(client net.Conn) (meta *proto.Meta, err error) {
	meta = &proto.Meta{
		Net: "tcp",
	}

	// read header
	header := make([]byte, 4)
	if _, err = io.ReadFull(client, header[:4]); err != nil {
		return
	}

	ver, cmd, _, atyp := header[0], header[1], header[2], header[3]
	if ver != Socks5Version {
		return nil, UnSupportVersion
	}

	// todo support other command
	if cmd != SocksCmdConnect {
		return nil, UnSupportCommand
	}

	var addr string

	// handle address
	switch atyp {
	case SocksIPv4Host:
		addrByte := make([]byte, 4)
		if _, err = io.ReadFull(client, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	case SocksIPv6Host:
		addrByte := make([]byte, 16)
		if _, err = io.ReadFull(client, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	case SocksDomainHost:
		addrLen := make([]byte, 1)
		if _, err = io.ReadFull(client, addrLen); err != nil {
			return
		}
		addrByte := make([]byte, int(addrLen[0]))
		if _, err = io.ReadFull(client, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	default:
		return nil, errors.New("unsupported address type")
	}

	// read the port
	portByte := make([]byte, 2)
	if _, err = io.ReadFull(client, portByte); err != nil {
		return nil, err
	}
	port := binary.BigEndian.Uint16(header[:2])

	meta.Address = fmt.Sprintf("%s:%d", addr, port)
	return
}
