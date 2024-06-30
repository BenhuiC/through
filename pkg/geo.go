package pkg

import (
	"context"
	"errors"
	"github.com/google/go-github/v62/github"
	"github.com/jasonlvhit/gocron"
	"github.com/oschwald/geoip2-golang"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"through/pkg/log"
	"time"
)

const (
	OWNER = "Loyalsoldier"
	REPO  = "geoip"
	FILE  = "Country.mmdb"
)

var db *Geo

func InitGeo(ctx context.Context, filepath string) (err error) {
	db, err = NewGeo(ctx, filepath)
	return
}

func Country(ip net.IP) string {
	return db.Country(ip)
}

type Geo struct {
	filepath  string
	db        *geoip2.Reader
	githubCli *github.Client
	sync.RWMutex
}

func NewGeo(ctx context.Context, filepath string) (g *Geo, err error) {
	if filepath == "" {
		filepath = FILE
	}
	g = &Geo{
		filepath:  filepath,
		githubCli: github.NewClient(nil),
		RWMutex:   sync.RWMutex{},
	}

	_, err = os.Stat(filepath)
	if errors.Is(err, os.ErrNotExist) {
		err = g.DownloadGeoDBFile()
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	err = g.LoadGeoDBFile()
	if err != nil {
		return nil, err
	}

	err = gocron.Every(1).Friday().Do(g.ReloadGeoDB)
	if err != nil {
		return
	}
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-gocron.Start():
			return
		}
	}()

	return
}

func (g *Geo) DownloadGeoDBFile() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.Info("downloading GeoDB file")
	release, _, err := g.githubCli.Repositories.GetLatestRelease(ctx, OWNER, REPO)
	if err != nil {
		return
	}
	dbUrl := ""
	for _, asset := range release.Assets {
		if *asset.Name == FILE {
			dbUrl = *asset.BrowserDownloadURL
			break
		}
	}
	if dbUrl == "" {
		log.Info("latest GeoDB file not found")
		return
	}

	resp, err := http.Get(dbUrl)
	if err != nil {
		log.Infof("failed to download GeoDB file: %v", err)
		return err
	}
	defer resp.Body.Close()

	fileInfo, err := os.Stat(g.filepath)
	if err == nil && fileInfo.Size() > int64(0) {
		_ = os.Remove(g.filepath)
	}

	file, err := os.Create(g.filepath)
	if err != nil {
		log.Infof("failed to open GeoDB file: %v", err)
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	return
}

func (g *Geo) LoadGeoDBFile() (err error) {
	g.Lock()
	defer g.Unlock()

	gdb, err := geoip2.Open(g.filepath)
	if err != nil {
		log.Errorf("load GeoDB file error: %s", err)
		return
	}
	if gdb != nil {
		g.db = gdb
	}
	return
}

func (g *Geo) ReloadGeoDB() {
	log.Info("reloading GeoDB")
	if err := g.DownloadGeoDBFile(); err != nil {
		log.Errorf("download GeoDB file error: %s", err)
		return
	}
	_ = g.LoadGeoDBFile()
}

func (g *Geo) Country(ip net.IP) string {
	g.RLock()
	defer g.RUnlock()
	c, _ := g.db.Country(ip)
	if c != nil {
		return c.Country.IsoCode
	}
	return ""
}
