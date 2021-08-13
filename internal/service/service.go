package service

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"liang/internal/dao"
	"liang/internal/model"

	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	cron3 "github.com/robfig/cron/v3"
)

var Provider = wire.NewSet(New)

// Service service.
type Service struct {
	ac        *paladin.Map
	dao       dao.Dao
	cron      *cron3.Cron
	netBwMap  map[string]int64 // 节点的网卡速度信息
	nodeNames []string
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	s.cron = cron3.New(cron3.WithSeconds())
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)

	var (
		hosts   []string
		netLoad []float64
	)
	if err = s.ac.Get("netbwMapKeys").Slice(&hosts); err != nil {
		log.Error("unmarshal config netbwMapKeys error: %v", err)
		return
	}
	s.nodeNames = hosts

	if err = s.ac.Get("netbwMapValues").Slice(&netLoad); err != nil {
		log.Error("get slice config netbwMapValues error: %v", err)
		return
	}
	keyLen, valueLen := len(hosts), len(netLoad)
	if keyLen != valueLen {
		err = fmt.Errorf("len of netbwMapKeys(%d) and netbwMapValues(%d) not euqal", keyLen, valueLen)
		return
	}

	netMap := make(map[string]int64)
	for i := 0; i < keyLen; i++ {
		// 内部计算单位统一为Kbit/s
		netmp := int64(netLoad[i] * (model.MbitPS / 1024))
		if netmp == 0 {
			err = fmt.Errorf("netload of %s is %f, should not be zero", hosts[i], netLoad[i])
		}
		netMap[hosts[i]] = netmp
	}
	s.netBwMap = netMap
	log.Info("netBwMap is %#v", netMap)

	// 同步prom状态信息
	var syncInterval string
	syncInterval, err = s.ac.Get("syncStatusInterval").String()
	if err != nil {
		log.Error("get syncStatusInterval from application.toml error: %v", err)
		return
	}
	// err = s.SyncNetIO()
	err = s.ParallelSyncInfo()
	if err != nil {
		return
	}
	// TODO: 做下判断，如果err次数过多，直接panic
	_, err = s.cron.AddFunc(syncInterval, func() {
		// innerErr := s.SyncNetIO()
		innerErr := s.ParallelSyncInfo()
		if innerErr != nil {
			log.Error("%v", innerErr)
			return
		}
		s.ParallelSyncInfo()
	})
	if err != nil {
		log.Error("add sync prom status error: %v", err)
		return
	}
	s.cron.Start()

	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}

// PromDemo demo of prometheus api
func (s *Service) PromDemo() {
	s.dao.RequestPromDemo()
}

func (s *Service) GetAllCache() (map[string](map[string]int64), error) {
	return s.dao.GetAllInfo()
}

// RequestPromInfo 从本地缓存获取diskIO/netIO/CPU Usage/Mem Usage信息
func (s *Service) RequestPromInfo() (map[string](map[string]int64), error) {
	var (
		netIO, diskIO, cpuUsage, memUsage map[string]int64
		err                               error
	)
	netIO, err = s.dao.RequestPromMaxNetIO()
	if err != nil {
		return nil, err
	}
	diskIO, err = s.dao.RequestPromMaxDiskIO()
	if err != nil {
		return nil, err
	}
	cpuUsage, err = s.dao.RequestPromCPUUsage()
	if err != nil {
		return nil, err
	}
	memUsage, err = s.dao.RequestPromMemUsage()
	if err != nil {
		return nil, err
	}

	return map[string](map[string]int64){
		"net_io":    netIO,
		"disk_io":   diskIO,
		"cpu_usage": cpuUsage,
		"mem_usage": memUsage,
	}, nil
}

func (s *Service) RequestPromNetIO(bwType string) (map[string]int64, error) {
	return s.dao.RequestPromNetIO(bwType)
}

func (s *Service) RequestPromDiskIO(diskType string) (map[string]int64, error) {
	return s.dao.RequestPromDiskIO(diskType)
}

func (s *Service) RequestPromMaxNetIO() (map[string]int64, error) {
	return s.dao.RequestPromMaxNetIO()
}

func (s *Service) RequestPromMaxDiskIO() (map[string]int64, error) {
	return s.dao.RequestPromMaxDiskIO()
}

func (s *Service) RequestPromCPUUsage() (map[string]int64, error) {
	return s.dao.RequestPromCPUUsage()
}

func (s *Service) RequestPromMemUsage() (map[string]int64, error) {
	return s.dao.RequestPromMemUsage()
}

// filterByNodeName 根据node name过滤结果
func (s *Service) filterByNodeName(inMap map[string]int64) map[string]int64 {
	nodeNames := s.nodeNames
	outMap := make(map[string]int64)
	for _, name := range nodeNames {
		if v, ok := inMap[name]; ok {
			outMap[name] = v
		}
	}

	if len(outMap) == 0 {
		log.Warn("value map from prometheus is %v, nodeName is %v, not match",
			inMap, s.nodeNames)
	}

	return outMap
}

func (s *Service) SyncNetIO() error {
	res, err := s.dao.RequestPromNetIO(model.NetIOTypeDown)
	if err != nil {
		log.Error("get netload from prom error: %v", err)
		return err
	}

	// 过滤掉不符合的nodes，prom可能监控不在k8s集群中的node
	netMap := s.filterByNodeName(res)
	if len(netMap) == 0 {
		return nil
	}

	err = s.dao.SetNetIO(netMap)
	if err != nil {
		log.Error("SetNetIO error: %v", err)
		return err
	}

	return nil
}

// ParallelSyncInfo 并发获取CPU/Mem/DiskIO/NetIO信息
func (s *Service) ParallelSyncInfo() error {
	type innerFunc func() (map[string]int64, error)
	funcArr := []innerFunc{s.dao.RequestPromMaxNetIO, s.dao.RequestPromMaxDiskIO, s.dao.RequestPromCPUUsage, s.dao.RequestPromMemUsage}

	funcKeyMap := make(map[string]string)
	var vkey string
	for _, ff := range funcArr {
		fName := runtime.FuncForPC(reflect.ValueOf(ff).Pointer()).Name()
		if strings.Contains(fName, "RequestPromMaxNetIO") {
			vkey = model.ResourceNetIOKey
		} else if strings.Contains(fName, "RequestPromMaxDiskIO") {
			vkey = model.ResourceDiskIOKey
		} else if strings.Contains(fName, "RequestPromCPUUsage") {
			vkey = model.ResourceCPUKey
		} else {
			vkey = model.ResourceMemKey
		}

		funcKeyMap[fName] = vkey
	}

	var wg sync.WaitGroup
	var returnErr error
	start := time.Now()
	for i := range funcArr {
		ff := funcArr[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			if returnErr != nil {
				return
			}
			fName := runtime.FuncForPC(reflect.ValueOf(ff).Pointer()).Name()
			resMap, err := ff()
			if err != nil {
				log.Error("[ParallelGetLoadInfo] %s error: %v", fName, err)
				returnErr = paladin.ErrDifferentTypes
				return
			}

			filtered := s.filterByNodeName(resMap)
			err = s.dao.SetKV(funcKeyMap[fName], filtered)
			if err != nil {
				returnErr = err
				log.Error("[ParallelGetLoadInfo] SetKV error: %v", err)
			}
		}()
	}

	wg.Wait()
	costTime := time.Now().Sub(start).String()
	log.Info("sync dynamic info costs %s", costTime)
	return returnErr
}
