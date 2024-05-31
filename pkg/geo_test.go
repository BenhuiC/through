package pkg

import (
	"context"
	"net"
	"os"
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
	if err := Init(context.Background(), "./Country.mmdb"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestCountry(t *testing.T) {
	ip := "8.136.83.38"
	i := net.IP{}
	if err := i.UnmarshalText([]byte(ip)); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s -> %s", ip, Country(i))

	time.Sleep(time.Hour)
}
