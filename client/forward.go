package client

import (
	"context"
	"crypto/tls"
	"errors"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"through/config"
	"through/log"
	"through/proto"
	"through/util"
	"time"
)

type Forward interface {
	Http(writer http.ResponseWriter, request *http.Request)
	Connect(conn net.Conn, meta *proto.Meta)
	Close()
}

type ForwardManger struct {
	forwardClients map[string]Forward
}

func NewForwardManger(ctx context.Context, server []config.ProxyServer, tlsCfg *tls.Config, poolSize int) (f *ForwardManger, err error) {
	f = &ForwardManger{forwardClients: map[string]Forward{}}
	if len(server) == 0 {
		err = errors.New("server config must more then zero")
		return
	}

	f.forwardClients["reject"] = &RejectClient{}
	f.forwardClients["direct"] = &DirectClient{}

	for _, c := range server {
		if _, ok := f.forwardClients[c.Name]; ok {
			continue
		}
		forwardCli := NewForwardClient(ctx, c.Net, c.Addr, poolSize, tlsCfg)
		f.forwardClients[c.Name] = forwardCli
	}
	return
}

func (f *ForwardManger) GetForward(name string) (forward Forward, ok bool) {
	forward, ok = f.forwardClients[name]
	return
}

func (f *ForwardManger) Close() {
	for _, v := range f.forwardClients {
		v.Close()
	}
}

// DirectClient no proxy, direct call request
type DirectClient struct{}

func (d *DirectClient) Http(writer http.ResponseWriter, request *http.Request) {
	removeProxyHeaders(request)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHTTPResponse(writer, resp)
}

func (d *DirectClient) Connect(conn net.Conn, meta *proto.Meta) {
	defer conn.Close()
	remote, err := net.Dial(meta.GetNet(), meta.GetAddress())
	if err != nil {
		log.Errorf("dial remote %v error", meta.GetAddress())
		return
	}

	util.CopyLoopWait(conn, remote)
}

func (d *DirectClient) Close() {}

// RejectClient reject request,for ad or black list
type RejectClient struct{}

func (r *RejectClient) Http(writer http.ResponseWriter, request *http.Request) {
	log.Infof("reject http request")
	http.Error(writer, "reject", http.StatusForbidden)
}

func (r *RejectClient) Connect(conn net.Conn, meta *proto.Meta) {
	defer conn.Close()
	_, _ = conn.Write([]byte("reject"))
	log.Infof("reject connect")
}

func (r *RejectClient) Close() {}

// ForwardClient forward request to target server
type ForwardClient struct {
	net    string
	addr   string
	pool   *ConnectionPool
	client *http.Client
	logger *log.Logger
}

func NewForwardClient(ctx context.Context, network, addr string, poolSize int, tlsCfg *tls.Config) (f *ForwardClient) {
	f = &ForwardClient{
		net:    network,
		addr:   addr,
		pool:   NewConnectionPool(ctx, tlsCfg, addr, poolSize),
		logger: log.NewLogger(zap.AddCallerSkip(1)).With("network", network).With("address", addr),
	}
	f.client = &http.Client{
		Transport: &http.Transport{
			DialContext: f.dialContext,
		},
	}

	return
}

func (f *ForwardClient) Http(writer http.ResponseWriter, request *http.Request) {
	removeProxyHeaders(request)
	resp, err := f.client.Do(request)
	if err != nil {
		f.logger.Errorf("do http request error: %v", err)
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHTTPResponse(writer, resp)
}

func (f *ForwardClient) Connect(conn net.Conn, meta *proto.Meta) {
	remote, err := f.dialContext(context.Background(), meta.GetNet(), meta.GetAddress())
	if err != nil {
		f.logger.Errorf("dial server error: %v", err)
		_ = conn.Close()
		return
	}

	util.CopyLoopWait(conn, remote)
}

func (f *ForwardClient) dialContext(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	timeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err = f.pool.Get(timeout)
	if err != nil {
		log.Errorf("%v get connection error %v", addr, err)
		return
	}

	meta := &proto.Meta{
		Net:     "tcp",
		Address: addr,
	}
	if err = proto.WriteMeta(conn, meta); err != nil {
		return
	}

	return conn, err
}

func (f *ForwardClient) Close() {
	if f.pool != nil {
		f.pool.Close()
	}
}

func copyHTTPResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func removeProxyHeaders(r *http.Request) {
	r.RequestURI = "" // this must be reset when serving a request with the client
	r.Header.Del("Accept-Encoding")
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	if r.Header.Get("Connection") == "close" {
		r.Close = false
	}
	r.Header.Del("Connection")
}
