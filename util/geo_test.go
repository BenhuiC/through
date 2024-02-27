package util

import (
	"net"
	"testing"
)

func TestCountry(t *testing.T) {
	ip := "8.136.83.38"
	i := net.IP{}
	if err := i.UnmarshalText([]byte(ip)); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s -> %s", ip, Country(i))
}
