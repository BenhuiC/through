package util

import (
	"net"
	"testing"
)

func TestCountry(t *testing.T) {
	i := net.IP{}
	if err := i.UnmarshalText([]byte("212.64.63.124")); err != nil {
		t.Fatal(err)
	}
	t.Logf("212.64.63.124 -> %s", Country(i))
}
