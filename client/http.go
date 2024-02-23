package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"through/proto"
	"time"
)

type HttpProxy struct {
	connPool    *ConnectionPool
	ruleManager *RuleManager
}

func (h *HttpProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// todo
}

func (h *HttpProxy) https(writer http.ResponseWriter, request *http.Request) {
	// todo
	return
}

func (h *HttpProxy) http(writer http.ResponseWriter, request *http.Request) {
	host := request.Host // todo maybe ip:port
	f := h.ruleManager.Get(host)
	if f == nil {
		responseError(writer, errors.New("rule not match"))
		return
	}

	f.Http(writer, request)
}

func NewHttpProxy(ctx context.Context, pool *ConnectionPool, rules []string) (p *HttpProxy, err error) {
	// todo
	return
}

type HttpHandler struct {
	connPool *ConnectionPool
}

func (h *HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// todo user ruler to mapping different client
	// todo server https connect method
	cli := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				timeout, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()

				conn, err := h.connPool.Get(timeout)
				if err != nil {
					return nil, err
				}

				meta := &proto.Meta{
					Net:     "tcp",
					Address: addr,
				}
				if err = proto.WriteMeta(conn, meta); err != nil {
					return nil, err
				}

				return conn, err
			},
		},
	}

	removeProxyHeaders(request)

	resp, err := cli.Do(request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	copyHTTPResponse(writer, resp)
}

func responseError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(err.Error()))
}
