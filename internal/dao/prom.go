package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/pkg/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

type PromDao struct {
	API v1.API
}

func NewPromDao(addr, user, pass string) (*PromDao, error) {
	client, err := api.NewClient(api.Config{
		Address: addr,
		// We can use amazing github.com/prometheus/common/config helper!
		RoundTripper: config.NewBasicAuthRoundTripper(user, config.Secret(pass), "", api.DefaultRoundTripper),
	})
	if err != nil {
		log.Error("Error creating client: %v\n", err)
		return nil, err
	}

	v1api := v1.NewAPI(client)
	promDao := new(PromDao)
	promDao.API = v1api

	return promDao, nil
}

// func (promDao *PromDao) ExecPromQL(promQL string) (error, model.Value) {
func (promDao *PromDao) ExecPromQL(promQL string) (error, model.Value) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := promDao.API.Query(ctx, promQL, time.Now())
	if err != nil {
		log.Error("Error querying Prometheus: %v", err)
		return err, nil
	}
	if len(warnings) > 0 {
		log.Error("Warnings: %v", warnings)
	}

	return nil, result
}

func (d *dao) QueryDemo() {
	// d.promDao.ExecPromQL("up")
	// d.promDao.ExecPromQL(`increase(node_network_receive_bytes_total{device=~"eth0"}[30s])`)
	d.promDao.ExecPromQL(`rate(node_network_receive_bytes_total{device=~"eth0"}[30s])*8/1024`)

}

func (d *dao) QueryBandwidth() (map[string]int64, error) {
	promQL := `rate(node_network_receive_bytes_total{device=~"eth0"}[30s])*8`
	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	vectorValue, ok := result.(model.Vector)
	if !ok {
		err := fmt.Errorf("type of result not %T, get %T", model.Vector{}, result)
		return nil, err
	}

	resMap := make(map[string]int64)
	for i := 0; i < len(vectorValue); i++ {
		tmp := vectorValue[i]
		resMap[string(tmp.Metric["job"])] = int64(tmp.Value * 10)
	}

	return resMap, err
}
