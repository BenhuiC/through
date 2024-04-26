package client

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"through/pkg/log"
	"through/pkg/proto"
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

	StatusSuccess            = 0x00
	StatusGenSocksFail       = 0x01
	StatusConnectNotAllow    = 0x02
	StatusNetworkUnReachable = 0x03
	StatusHostUnReachable    = 0x04
	StatusConnectRefuse      = 0x05
	StatusTTLExpire          = 0x06
	StatusCommandNotSupport  = 0x07
	StatusAddressNotSupport  = 0x08
)

var (
	UnSupportVersion = errors.New("unsupported socks version")
	UnSupportCommand = errors.New("unsupported command")
)

// SocksProxy socks5 proxy
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
	go func() {
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
		log.Infof("socks host %v math server %v", meta.GetAddress(), server)

		f.Connect(conn, meta)
	}()
}

func (s *SocksProxy) readMetaFromConn(conn net.Conn) (meta *proto.Meta, err error) {
	if err = s.auth(conn); err != nil {
		return
	}
	return s.connect(conn)
}

func (s *SocksProxy) auth(conn net.Conn) (err error) {
	/*
		+----+----------+----------+
		|VER | NMETHODS | METHODS  |
		+----+----------+----------+
		| 1  |    1     | 1 to 255 |
		+----+----------+----------+
	*/
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
	/*
		+----+--------+
		|VER | METHOD |
		+----+--------+
		| 1  |   1    |
		+----+--------+
	*/
	_, err = conn.Write([]byte{Socks5Version, SocksNoAuthentication})

	// todo support auth
	/*
		+----+------+----------+------+----------+
		|VER | ULEN |  UNAME   | PLEN |  PASSWD  |
		+----+------+----------+------+----------+
		| 1  |  1   | 1 to 255 |  1   | 1 to 255 |
		+----+------+----------+------+----------+
	*/

	return
}

func (s *SocksProxy) connect(conn net.Conn) (meta *proto.Meta, err error) {
	meta = &proto.Meta{
		Net: "tcp",
	}
	/*
		+----+-----+-------+------+----------+----------+
		|VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
		VER 版本号，socks5的值为0x05
		CMD
		0x01表示CONNECT请求
		0x02表示BIND请求
		0x03表示UDP转发
		RSV 保留字段，值为0x00
		ATYP 目标地址类型，DST.ADDR的数据对应这个字段的类型。
		0x01表示IPv4地址，DST.ADDR为4个字节
		0x03表示域名，DST.ADDR是一个可变长度的域名
		0x04表示IPv6地址，DST.ADDR为16个字节长度
		DST.ADDR 一个可变长度的值
		DST.PORT 目标端口，固定2个字节
	*/

	// read header
	header := make([]byte, 4)
	if _, err = io.ReadFull(conn, header[:4]); err != nil {
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
		if _, err = io.ReadFull(conn, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	case SocksIPv6Host:
		addrByte := make([]byte, 16)
		if _, err = io.ReadFull(conn, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	case SocksDomainHost:
		addrLen := make([]byte, 1)
		if _, err = io.ReadFull(conn, addrLen); err != nil {
			return
		}
		addrByte := make([]byte, int(addrLen[0]))
		if _, err = io.ReadFull(conn, addrByte); err != nil {
			return
		}
		addr = string(addrByte)
	default:
		return nil, errors.New("unsupported address type")
	}

	// read the port
	portByte := make([]byte, 2)
	if _, err = io.ReadFull(conn, portByte); err != nil {
		return nil, err
	}
	port := binary.BigEndian.Uint16(portByte)

	meta.Address = fmt.Sprintf("%s:%d", addr, port)

	/*
		+----+-----+-------+------+----------+----------+
		|VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		+----+-----+-------+------+----------+----------+
		| 1  |  1  | X'00' |  1   | Variable |    2     |
		+----+-----+-------+------+----------+----------+
		VER socks版本，这里为0x05
		REP Relay field,内容取值如下
		X’00’ succeeded
		X’01’ general SOCKS server failure
		X’02’ connection not allowed by ruleset
		X’03’ Network unreachable
		X’04’ Host unreachable
		X’05’ Connection refused
		X’06’ TTL expired
		X’07’ Command not supported
		X’08’ Address type not supported
		X’09’ to X’FF’ unassigned
		RSV 保留字段
		ATYPE 同请求的ATYPE
		BND.ADDR 服务绑定的地址
		BND.PORT 服务绑定的端口DST.PORT
	*/

	// write response
	resp := []byte{Socks5Version, StatusSuccess, 0x00}
	resp = append(resp, s.parseAddr(conn.LocalAddr().String())...)
	_, err = conn.Write(resp)
	if err != nil {
		return nil, errors.New("write rsp: " + err.Error())
	}

	return
}

// parseAddr parses the address in string s. Returns nil if failed.
func (s *SocksProxy) parseAddr(str string) (addr []byte) {
	def := []byte{SocksIPv4Host, 0, 0, 0, 0, 0, 0}
	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return def
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = SocksIPv4Host
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = SocksIPv6Host
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return def
		}
		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = SocksDomainHost
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return def
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portnum>>8), byte(portnum)

	return addr
}
