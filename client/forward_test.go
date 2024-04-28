package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"through/util"
	"time"
)

func TestForwardClient_Connect(t *testing.T) {
	tlsCfg, err := util.LoadTlsConfig("../cert/client.key", "../cert/client.crt", "", true)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	forwardCli := NewForwardClient(ctx, "tcp", "xxx", 10, tlsCfg)

	requestCnt := 30
	var totalCost int64
	for i := 0; i < requestCnt; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://google.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		st := time.Now()
		resp, err := forwardCli.client.Do(req)
		totalCost += time.Since(st).Milliseconds()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			t.Fatal(fmt.Sprintf("error response with code %d", resp.StatusCode))
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	t.Logf("%d request totalCost cost %d Milliseconds, avg %d Milliseconds", requestCnt, totalCost, totalCost/int64(requestCnt))
}

func TestGrpcForwardClient_Connect(t *testing.T) {
	tlsCfg, err := util.LoadTlsConfig("../cert/client.key", "../cert/client.crt", "", true)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	forwardCli := NewGrpcForwardClient(ctx, "tcp", "xxx", 10, tlsCfg)

	requestCnt := 30
	var totalCost int64
	for i := 0; i < requestCnt; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://google.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		st := time.Now()
		resp, err := forwardCli.client.Do(req)
		totalCost += time.Since(st).Milliseconds()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			t.Fatal(fmt.Sprintf("error response with code %d", resp.StatusCode))
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	t.Logf("%d request totalCost cost %d Milliseconds, avg %d Milliseconds", requestCnt, totalCost, totalCost/int64(requestCnt))
}
