package client

import (
	"context"
	"sync"
	"testing"
	"through/config"
	"through/pkg/log"
	"time"
)

func TestMain(m *testing.M) {
	config.Common = &config.CommonCfg{
		Env:     "prod",
		LogFile: "",
	}
	if err := log.Init(); err != nil {
		panic(err)
	}
	m.Run()
}

func TestResolverManager_Lookup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	r, err := NewResolverManger(ctx, []config.ResolverServer{
		{DoT: "185.222.222.222"},
		{DoT: "dns.pub"},
		{DoT: "223.6.6.6"},
	})
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			t.Logf("lookup www.baidu.com -> %v", r.Lookup("www.baidu.com"))
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			t.Logf("lookup www.google.com -> %v", r.Lookup("www.google.com"))
		}()
	}

	wg.Wait()
	t.Logf("total cose %v", time.Now().Sub(start))
}
