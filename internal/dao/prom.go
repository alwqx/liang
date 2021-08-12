package dao

import (
	"context"
	"fmt"
	"math"
	"time"

	liangModel "liang/internal/model"

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

func (d *dao) RequestPromDemo() {
	// d.promDao.ExecPromQL("up")
	// d.promDao.ExecPromQL(`increase(node_network_receive_bytes_total{device=~"eth0"}[30s])`)
	d.promDao.ExecPromQL(`max(irate(node_network_receive_bytes_total[30s])*8/1024) by (job)`)

}

func (d *dao) parsePromResultInt64(result model.Value, base int) (map[string]int64, error) {
	vectorValue, ok := result.(model.Vector)
	if !ok {
		err := fmt.Errorf("type of result not %T, get %T", model.Vector{}, result)
		return nil, err
	}

	resMap := make(map[string]int64)
	for i := 0; i < len(vectorValue); i++ {
		tmp := vectorValue[i]
		tmpv := float64(tmp.Value)
		if base > 1 {
			resMap[string(tmp.Metric["job"])] = int64(math.Round(tmpv * float64(base)))
		} else {
			resMap[string(tmp.Metric["job"])] = int64(math.Round(tmpv))
		}
	}

	return resMap, nil
}

func (d *dao) parsePromResultFloat64(result model.Value) (map[string]float64, error) {
	vectorValue, ok := result.(model.Vector)
	if !ok {
		err := fmt.Errorf("type of result not %T, get %T", model.Vector{}, result)
		return nil, err
	}

	resMap := make(map[string]float64)
	for i := 0; i < len(vectorValue); i++ {
		tmp := vectorValue[i]
		resMap[string(tmp.Metric["job"])] = float64(tmp.Value)
	}

	return resMap, nil
}

// RequestPromNetIO 获取网络负载，根据参数决定是下载负载还是上传负载
// 单位 kbit/s
func (d *dao) RequestPromNetIO(bwType string) (map[string]int64, error) {
	var promQL string
	if bwType == liangModel.NetIOTypeDown {
		promQL = `max(irate(node_network_receive_bytes_total[30s])*8/1000) by (job)`
	} else {
		promQL = `max(irate(node_network_transmit_bytes_total[30s])*8/1000) by (job)`
	}
	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 1)
}

// RequestPromMaxNetIO 查询上行/下行中最大网络IO
// 单位 kbit/s
func (d *dao) RequestPromMaxNetIO() (map[string]int64, error) {
	promQL := `(max(irate(node_network_receive_bytes_total[30s])*8/1000) by (job)) > (max(irate(node_network_transmit_bytes_total[30s])*8/1024) by (job)) or (max(irate(node_network_transmit_bytes_total[30s])*8/1024) by (job))`
	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 1)
}

// RequestPromDiskIO 查询Prom上机器的DiskIO
// 单位byte/s 或者 B/s
func (d *dao) RequestPromDiskIO(diskType string) (map[string]int64, error) {
	var promQL string
	if diskType == liangModel.DiskIOTypeWrite {
		promQL = `max(irate(node_disk_written_bytes_total[30s])) by (job)`
	} else {
		promQL = `max(irate(node_disk_read_bytes_total[30s])) by (job)`
	}

	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 1)
}

// RequestPromMaxDiskIO 查询读/写中最大磁盘IO
func (d *dao) RequestPromMaxDiskIO() (map[string]int64, error) {
	promQL := `(max(irate(node_disk_written_bytes_total[30s])) by (job)) > (max(irate(node_disk_read_bytes_total[30s])) by (job)) or (max(irate(node_disk_read_bytes_total[30s])) by (job))`
	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 1)
}

// RequestPromCPUUsage 查询Prom上机器的CPU使用率
// 取4位有效数字后转换成int64，相比float64满足精度的前提下提高计算速度
// e.g.: 0.012->12 23.453453245->2345
func (d *dao) RequestPromCPUUsage() (map[string]int64, error) {
	promQL := `(1 - avg(rate(node_cpu_seconds_total{mode="idle"}[30s])) by (job))`

	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 100)
}

// RequestPromMemUsage 查询Prom上机器的内存使用率
// 取4位有效数字后转换成int64，相比float64满足精度的前提下提高计算速度
// e.g.: 0.012->12 23.453453245->2345
func (d *dao) RequestPromMemUsage() (map[string]int64, error) {
	promQL := `(1 - (node_memory_MemAvailable_bytes / (node_memory_MemTotal_bytes)))`

	err, result := d.promDao.ExecPromQL(promQL)
	if err != nil {
		return nil, err
	}

	return d.parsePromResultInt64(result, 100)
}
