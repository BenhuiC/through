package util

import (
	"bytes"
	_ "embed"
	"io"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/xi2/xz"
)

//go:embed Country.mmdb.xz
var dbBytes []byte

var db *geoip2.Reader

func init() {
	r, err := xz.NewReader(bytes.NewReader(dbBytes), 0)
	if err != nil {
		panic(err)
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	db, err = geoip2.FromBytes(raw)
	if err != nil {
		panic(err)
	}
}

func Country(ip net.IP) string {
	c, _ := db.Country(ip)
	if c != nil {
		return c.Country.IsoCode
	}
	return ""
}
