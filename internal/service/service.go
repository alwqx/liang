package service

import (
	"context"
	"fmt"

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
	err = s.SyncNetload()
	if err != nil {
		return
	}
	// TODO: 做下判断，如果err次数过多，直接panic
	_, err = s.cron.AddFunc(syncInterval, func() {
		innerErr := s.SyncNetload()
		if innerErr != nil {
			log.Error("%v", innerErr)
			return
		}
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
	s.dao.QueryDemo()
}

func (s *Service) QueryBandwidth() (map[string]int64, error) {
	return s.dao.QueryBandwidth()
}

func (s *Service) SyncNetload() error {
	res, err := s.dao.QueryBandwidth()
	if err != nil {
		log.Error("get netload from prom error: %v", err)
		return err
	}

	// 过滤掉不符合的nodes，prom可能监控不在k8s集群中的node
	nodeNames := s.nodeNames
	netMap := make(map[string]int64)
	for _, name := range nodeNames {
		if netCap, ok := res[name]; ok {
			netMap[name] = netCap
		}
	}
	if len(netMap) == 0 {
		log.Error("netload from prometheus is %v, nodeName is %v, not match",
			res, nodeNames)
		return nil
	}

	err = s.dao.SetNetload(netMap)
	if err != nil {
		log.Error("SetNetload error: %v", err)
		return err
	}

	return nil
}
