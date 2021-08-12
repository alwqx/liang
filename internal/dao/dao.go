package dao

import (
	"context"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

//go:generate kratos tool genbts
// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	// prometheus related interface
	RequestPromDemo()
	RequestPromMaxDiskIO() (map[string]int64, error)
	RequestPromMaxNetIO() (map[string]int64, error)
	RequestPromNetIO(bwType string) (map[string]int64, error)
	RequestPromDiskIO(diskType string) (map[string]int64, error)
	RequestPromCPUUsage() (map[string]int64, error)
	RequestPromMemUsage() (map[string]int64, error)

	// local KV cache interface
	SetKV(k string, v interface{}) error
	GetAllInfo() (map[string](map[string]int64), error)
	SetNetIO(netload map[string]int64) error
	GetNetIO() (map[string]int64, error)
}

// dao dao.
type dao struct {
	promDao    *PromDao
	localCache gcache.Cache
	demoExpire int32
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	var cfg struct {
		PromAddr              string
		PromBasicAuthUser     string
		PromBasicAuthPassword string
		LocalCacheExpire      int64
		DemoExpire            xtime.Duration
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}

	// new promDao
	promDao, err := NewPromDao(cfg.PromAddr, cfg.PromBasicAuthUser, cfg.PromBasicAuthPassword)
	if err != nil {
		return
	}

	d = &dao{
		promDao:    promDao,
		localCache: gcache.New(2000).LRU().Expiration(time.Duration(cfg.LocalCacheExpire) * time.Second).Build(),
		demoExpire: int32(time.Duration(cfg.DemoExpire) / time.Second),
	}
	cf = d.Close

	return
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
