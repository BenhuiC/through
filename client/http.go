package client

import (
	"context"
	"io"
	"net"
	"net/http"
	"through/proto"
	"time"
)

var (
	ProxyHeaders = map[string]bool{
		"Proxy-Connection":    true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Connection":          true,
	}
)

type HttpHandler struct {
	connPool *ConnectionPool
}

func (h *HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// todo user ruler to mapping different client
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

	h.processRequest(request)

	resp, err := cli.Do(request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	CopyHTTPResponse(writer, resp)
}

func CopyHTTPResponse(w http.ResponseWriter, resp *http.Response) {
	for k, v := range resp.Header {
		w.Header().Set(k, v[0])
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *HttpHandler) processRequest(req *http.Request) {
	// todo
	req.RequestURI = "" // this must be reset when serving a request with the client
	// req.Header.Del("Accept-Encoding")
	for k := range ProxyHeaders {
		req.Header.Del(k)
	}
}
