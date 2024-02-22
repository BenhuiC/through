package client

import (
	"net/http"
)

type HttpHandler struct {
}

func (h HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// todo
}
