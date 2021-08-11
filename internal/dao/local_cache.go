package dao

import (
	"errors"
	"fmt"
	"liang/internal/model"

	"github.com/bluele/gcache"
	"github.com/go-kratos/kratos/pkg/log"
)

func (d *dao) innerGet(key string) (map[string]int64, error) {
	rpl, err := d.localCache.Get(key)
	if err != nil {
		if !errors.Is(err, gcache.KeyNotFoundError) {
			log.Error("get value of key %s from local cache error: %v", key, err)
			return nil, err
		}
		return nil, nil
	}

	if rpl == nil {
		log.Info("get empty value of key %s from local cache", key)
		return nil, nil
	}

	res, ok := rpl.(map[string]int64)
	if !ok {
		err = fmt.Errorf("value of key %s is not type of map[string]int64", key)
		log.Error("%v", err)
		return nil, err
	}

	return res, nil
}

func (d *dao) SetKV(k string, v interface{}) error {
	return d.localCache.Set(k, v)
}

func (d *dao) SetNetIO(netIO map[string]int64) error {
	return d.localCache.Set(model.ResourceNetIOKey, netIO)
}

func (d *dao) GetNetIO() (map[string]int64, error) {
	return d.innerGet(model.ResourceNetIOKey)
}

func (d *dao) SetDiskIO(diskIO map[string]int64) error {
	return d.localCache.Set(model.ResourceDiskIOKey, diskIO)
}

func (d *dao) GetDiskIO() (map[string]int64, error) {
	return d.innerGet(model.ResourceDiskIOKey)
}

func (d *dao) SetCPUUsage(cpuUsage map[string]int64) error {
	return d.localCache.Set(model.ResourceCPUKey, cpuUsage)
}

func (d *dao) GetCPUUsage() (map[string]int64, error) {
	return d.innerGet(model.ResourceCPUKey)
}

func (d *dao) SetMemUsage(memUsage map[string]int64) error {
	return d.localCache.Set(model.ResourceMemKey, memUsage)
}

func (d *dao) GetMemUsage() (map[string]int64, error) {
	return d.innerGet(model.ResourceMemKey)
}
