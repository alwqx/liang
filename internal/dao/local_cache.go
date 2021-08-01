package dao

import (
	"errors"
	"fmt"
	"liang/internal/model"

	"github.com/bluele/gcache"
	"github.com/go-kratos/kratos/pkg/log"
)

func (d *dao) SetNetload(netload map[string]int64) error {
	err := d.localCache.Set(model.ResourceNetloadKey, netload)
	return err
}

func (d *dao) GetNetload() (map[string]int64, error) {
	rpl, err := d.localCache.Get(model.ResourceNetloadKey)
	if err != nil {
		if !errors.Is(err, gcache.KeyNotFoundError) {
			log.Error("get value of key %s from local cache error: %v", model.ResourceNetloadKey, err)
			return nil, err
		}
		return nil, nil
	}

	if rpl == nil {
		log.Info("get empty value of key %s from local cache", model.ResourceNetloadKey)
		return nil, nil
	}

	res, ok := rpl.(map[string]int64)
	if !ok {
		err = fmt.Errorf("value of key %s is not type of map[string]int64", model.ResourceNetloadKey)
		log.Error("%v", err)
		return nil, err
	}

	return res, nil
}
