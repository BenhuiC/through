package pkg

import (
	"github.com/dustin/go-humanize"
	"sync/atomic"
	"through/pkg/log"
	"time"
)

var Metrics *metrics

func InitMetrics() {
	Metrics = &metrics{
		logger: log.NewLogger(),
	}
	go Metrics.doPrintMetrics()
}

type metrics struct {
	upload       atomic.Uint64
	download     atomic.Uint64
	lastUpload   uint64
	lastDownload uint64
	logger       *log.Logger
}

func (m *metrics) Upload(dataLen uint64) {
	m.upload.Add(dataLen)
}

func (m *metrics) Download(dataLen uint64) {
	m.download.Add(dataLen)
}

func (m *metrics) doPrintMetrics() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		curUpload := m.upload.Load()
		curDownload := m.download.Load()
		m.logger.With("upload", humanize.Bytes(curUpload-m.lastUpload)).With("download", humanize.Bytes(curDownload-m.lastDownload)).Info("traffic metrics")
		m.lastUpload = curUpload
		m.lastDownload = curDownload
	}
}
