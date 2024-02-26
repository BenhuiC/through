package client

import (
	"testing"
	"through/config"
)

func TestResolverManager_Lookup(t *testing.T) {
	r, err := NewResolverManger([]config.ResolverServer{
		{DoT: "185.222.222.222"},
		{DoT: "dns.pub"},
		{DoT: "223.6.6.6"},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("lookup www.baidu.com -> %v", r.Lookup("www.baidu.com"))
	t.Logf("lookup www.google.com -> %v", r.Lookup("www.google.com"))
}
