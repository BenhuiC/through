package client

import (
	"context"
	"net/http"
	"through/log"
	"through/proto"
)

type HttpProxy struct {
	forwardManager *ForwardManger
	ruleManager    *RuleManager
}

func NewHttpProxy(ctx context.Context, forwards *ForwardManger, rules *RuleManager) (p *HttpProxy) {
	p = &HttpProxy{
		forwardManager: forwards,
		ruleManager:    rules,
	}

	return
}

func (h *HttpProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Infof("http proxy host: %v, method: %v", request.Host, request.Method)
	if request.Method == http.MethodConnect {
		h.https(writer, request)
	} else {
		h.http(writer, request)
	}
}

func (h *HttpProxy) https(writer http.ResponseWriter, request *http.Request) {
	hij, ok := writer.(http.Hijacker)
	if !ok {
		log.Errorf("httpserver does not support hijacking")
		http.Error(writer, "httpserver does not support hijacking", http.StatusServiceUnavailable)
		return
	}

	host := request.URL.Host
	server := h.ruleManager.Get(host)
	f, ok := h.forwardManager.GetForward(server)
	if !ok {
		log.Infof("host %v math no server", host)
		http.Error(writer, "rule match no server", http.StatusServiceUnavailable)
		return
	}
	log.Infof("https host %v math server %v", host, server)

	proxyClient, _, e := hij.Hijack()
	if e != nil {
		log.Infof("cannot hijack connection %v", e)
		http.Error(writer, "cannot hijack connection "+e.Error(), http.StatusServiceUnavailable)
		return
	}

	_, _ = proxyClient.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))

	f.Connect(proxyClient, &proto.Meta{Net: "tcp", Address: request.URL.Host})
	return
}

func (h *HttpProxy) http(writer http.ResponseWriter, request *http.Request) {
	if !request.URL.IsAbs() {
		http.Error(writer, "This is a proxy server. Does not respond to non-proxy requests.", http.StatusBadRequest)
		return
	}
	host := request.URL.Host
	server := h.ruleManager.Get(host)
	f, ok := h.forwardManager.GetForward(server)
	if !ok {
		log.Infof("host %v math no server", host)
		http.Error(writer, "rule match no server", http.StatusServiceUnavailable)
		return
	}
	log.Infof("http host %v math server %v", host, server)

	f.Http(writer, request)
}
