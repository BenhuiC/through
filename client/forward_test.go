package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
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

	wg := sync.WaitGroup{}
	requests := loadHttps()
	reqChan := make(chan *TestHttp, 10)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range requests {
			reqChan <- &requests[i]
		}
		close(reqChan)
	}()

	curr := 10
	for i := 0; i < curr; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var totalCost int64
			var reqCnt int
			for {
				req := <-reqChan
				if req == nil {
					break
				}
				httpReq, err := http.NewRequest(req.Method, req.Url, nil)
				if err != nil {
					t.Fatal(err)
				}
				st := time.Now()
				resp, err := forwardCli.client.Do(httpReq)
				cost := time.Since(st).Milliseconds()
				reqCnt++
				totalCost += cost
				if err != nil {
					t.Fatal(err)
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			t.Logf("goroutine %d do %d request totalCost cost %d Milliseconds, avg %d Milliseconds", idx, reqCnt, totalCost, totalCost/int64(reqCnt))
		}(i)
	}

	wg.Wait()
}

func TestGrpcForwardClient_Connect(t *testing.T) {
	tlsCfg, err := util.LoadTlsConfig("../cert/client.key", "../cert/client.crt", "", true)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	forwardCli := NewGrpcForwardClient(ctx, "tcp", "xxx", 10, tlsCfg)

	wg := sync.WaitGroup{}
	requests := loadHttps()
	reqChan := make(chan *TestHttp, 10)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range requests {
			reqChan <- &requests[i]
		}
		close(reqChan)
	}()

	curr := 10
	for i := 0; i < curr; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var totalCost int64
			var reqCnt int
			for {
				req := <-reqChan
				if req == nil {
					break
				}
				httpReq, err := http.NewRequest(req.Method, req.Url, nil)
				if err != nil {
					t.Fatal(err)
				}
				st := time.Now()
				resp, err := forwardCli.client.Do(httpReq)
				cost := time.Since(st).Milliseconds()
				reqCnt++
				totalCost += cost
				if err != nil {
					t.Fatal(err)
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			t.Logf("goroutine %d do %d request totalCost cost %d Milliseconds, avg %d Milliseconds", idx, reqCnt, totalCost, totalCost/int64(reqCnt))
		}(i)
	}

	wg.Wait()
}

type TestHttp struct {
	Method string `json:"method"`
	Url    string `json:"url"`
}

func loadHttps() []TestHttp {
	res := make([]TestHttp, 0)
	// your path
	data, err := os.ReadFile("../.tmp/test_1000.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &res); err != nil {
		panic(err)
	}

	return res
}
