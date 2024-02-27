package client

import (
	"context"
	"testing"
	"through/config"
	"through/log"
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

	t.Logf("lookup www.baidu.com -> %v", r.Lookup("www.baidu.com"))
	t.Logf("lookup www.google.com -> %v", r.Lookup("www.google.com"))

	time.Sleep(32 * time.Second)
	t.Logf("lookup www.baidu.com -> %v", r.Lookup("www.baidu.com"))
	t.Logf("lookup www.google.com -> %v", r.Lookup("www.google.com"))
}
