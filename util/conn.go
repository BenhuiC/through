package util

import (
	"io"
	"net"
	"sync"
)

func CopyLoopWait(c1 net.Conn, c2 net.Conn) {
	var wg sync.WaitGroup
	cp := func(dst, src net.Conn) {
		defer wg.Done()
		_, err := io.Copy(dst, src)
		_ = dst.Close()
		if err != nil {
			_ = src.Close()
		}
	}
	wg.Add(2)
	go cp(c1, c2)
	go cp(c2, c1)
	wg.Wait()
}
