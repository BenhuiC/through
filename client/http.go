package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"through/config"
	"through/log"
	"through/proto"
)

type HttpProxy struct {
	forwards    map[string]Forward
	ruleManager *RuleManager
}

func (h *HttpProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Info("http proxy host: %v, method: %v", request.Host, request.Method)
	if request.Method == http.MethodConnect {
		h.https(writer, request)
	} else {
		h.http(writer, request)
	}
}

func (h *HttpProxy) https(writer http.ResponseWriter, request *http.Request) {
	hij, ok := writer.(http.Hijacker)
	if !ok {
		log.Info("httpserver does not support hijacking")
		http.Error(writer, "httpserver does not support hijacking", http.StatusServiceUnavailable)
		return
	}

	proxyClient, _, e := hij.Hijack()
	if e != nil {
		log.Info("cannot hijack connection %v", e)
		http.Error(writer, "cannot hijack connection "+e.Error(), http.StatusServiceUnavailable)
		return
	}

	_, _ = proxyClient.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))

	host := request.Host
	server := h.ruleManager.Get(host)
	f, ok := h.forwards[server]
	if !ok {
		log.Info("host %v math no server", host)
		http.Error(writer, "rule match no server", http.StatusServiceUnavailable)
		return
	}

	f.Connect(proxyClient, &proto.Meta{Net: "tcp", Address: request.URL.Host})
	return
}

func (h *HttpProxy) http(writer http.ResponseWriter, request *http.Request) {
	if !request.URL.IsAbs() {
		http.Error(writer, "This is a proxy server. Does not respond to non-proxy requests.", http.StatusBadRequest)
		return
	}
	host := request.Host
	server := h.ruleManager.Get(host)
	f, ok := h.forwards[server]
	if !ok {
		log.Info("host %v math no server", host)
		http.Error(writer, "rule match no server", http.StatusServiceUnavailable)
		return
	}

	f.Http(writer, request)
}

func (h *HttpProxy) Close() {
	for _, f := range h.forwards {
		f.Close()
	}
}

func NewHttpProxy(ctx context.Context, tlsCfg *tls.Config, poolSize int, server []config.ProxyServer, rules []string) (p *HttpProxy, err error) {
	if len(server) == 0 {
		err = errors.New("server config must more then zero")
		return
	}
	p = &HttpProxy{}
	p.ruleManager, err = NewRuleManager(rules)
	if err != nil {
		return
	}

	p.forwards = make(map[string]Forward)

	p.forwards["reject"] = &RejectClient{}
	p.forwards["direct"] = &DirectClient{}

	for _, c := range server {
		if _, ok := p.forwards[c.Name]; ok {
			continue
		}
		forwardCli := NewForwardClient(ctx, c.Net, c.Addr, poolSize, tlsCfg)
		p.forwards[c.Name] = forwardCli
	}

	return
}
