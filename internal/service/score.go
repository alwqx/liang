package service

import (
	"liang/internal/model"

	"github.com/go-kratos/kratos/pkg/log"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (s *Service) Prioritize(args *extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	if s.useBNP {
		log.V(3).Info("use bnp algo to score...")
		return s.bnpScore(args)
	}

	log.V(3).Info("use cmdn topsis algo to score...")
	return s.cmdapScore(args)
}

func (s *Service) bnpScore(args *extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	curMap, err := s.dao.GetNetIO()
	if err != nil || len(curMap) == 0 {
		log.Error("Prioritize: get empty curMap %v or run into error: %v",
			curMap, err)
		return nil, err
	}

	bnp := BalanceNetloadPriority{}
	res, err := bnp.Score(args.Pod, *args.NodeNames, curMap, s.netBwMap)
	log.V(3).Info("score result of BNP is: %#v", res)

	return &res, err
}

// cmdapScore cmdap算法评分入口
func (s *Service) cmdapScore(args *extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	nodeNames := *args.NodeNames
	cacheData, err := s.GetAllCache()
	if err != nil {
		log.Error("get all cache data error: %v", err)
		return nil, err
	}

	cmdn := CMDNPriority{}
	res, err := cmdn.Score(args.Pod, nodeNames, s.netBwMap, cacheData)
	log.V(3).Info("score result of CMDAP is: %#v", res)
	if err == nil && s.topsisMin {
		for i := range res {
			res[i].Score = model.MaxNodeScore - res[i].Score
		}
	}

	return &res, err
}
