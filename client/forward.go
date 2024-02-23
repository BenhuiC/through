package client

import (
	"io"
	"net/http"
)

type Forward interface {
	Http(writer http.ResponseWriter, request *http.Request)
}

type DirectClient struct {
}

func (d *DirectClient) Http(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

type RejectClient struct{}

func (r *RejectClient) Http(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

type ForwardClient struct {
	net  string
	addr string
}

func (f *ForwardClient) Http(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
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
