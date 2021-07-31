package dao

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/sync/pipeline/fanout"
	xtime "github.com/go-kratos/kratos/pkg/time"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

//go:generate kratos tool genbts
// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)

	QueryDemo()
	QueryBandwidth() (error, map[string]int64)
}

// dao dao.
type dao struct {
	promDao    *PromDao
	cache      *fanout.Fanout
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
		cache:      fanout.New("cache"),
		demoExpire: int32(time.Duration(cfg.DemoExpire) / time.Second),
	}
	cf = d.Close

	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
