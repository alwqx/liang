package service

import (
	"strconv"
	"strings"

	"liang/internal/model"

	"github.com/go-kratos/kratos/pkg/log"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	v1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func (s *Service) Prioritize(args *extenderv1.ExtenderArgs) *extenderv1.HostPriorityList {
	bnp := BalanceNetloadPriority{}
	// 获取curMap
	curMap, err := s.dao.GetNetload()
	if err != nil || len(curMap) == 0 {
		log.Error("Prioritize: get empty curMap %v or run into error: %v",
			curMap, err)
		return nil
	}
	res := bnp.Score(args.Pod, *args.NodeNames, curMap, s.netBwMap)
	log.V(3).Info("score result of BNP is: %#v", res)

	return &res
}

type BalanceNetloadPriority struct{}

// Score Node评分算法
// 需要的Node动态资源信息已经在ExtendResource中提供了，Score算法要结合Pod中的资源请求
// 进行计算，然后根据结果对Node打分，注意这里的打分不需要加上权重，权重有scheduler根据
// HTTPExtender统一分配一个直接的权重
// 动态可压缩资源在Pod.MetaData的Annotation中以map形式定义
func (algo *BalanceNetloadPriority) Score(pod *v1.Pod, nodeNames []string, curMap map[string]int64, capMap map[string]int64) extenderv1.HostPriorityList {
	neededMap := make(map[string]int64)
	for k, v := range pod.Annotations {
		if !strings.Contains(k, model.PodAnnotationKey) {
			continue
		}
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Error("parse %s to int64 error:%v", v, err)
			continue
		}

		neededMap[k] = vInt
	}

	netNeed := neededMap[model.ResourceNetloadKey] * model.KbitPS
	validNames := make([]string, 0)
	curArr := make([]float64, 0)
	capArr := make([]float64, 0)

	nodeNum := len(nodeNames)
	for i := 0; i < nodeNum; i++ {
		nodeName := nodeNames[i]
		netCur, ok := curMap[nodeName]
		if !ok {
			log.Warn("current net info of node %s does not exist, skip", nodeName)
			continue
		}

		netCap, ok1 := capMap[nodeName]
		// 过滤掉不存在或者资源超出的情况
		if !ok1 {
			log.Warn("cap net info of node %s does not exist, skip", nodeName)
			continue
		}

		if netCur+netNeed > netCap {
			log.Warn("request net %d plus cur net %d overflow net cap %d, skip",
				netNeed, netCur, netCap)
			continue
		}

		validNames = append(validNames, nodeName)
		curArr = append(curArr, float64(curMap[nodeName]))
		capArr = append(capArr, float64(capMap[nodeName]))
	}

	// 没有一个node符合条件
	if len(validNames) == 0 {
		log.Warn("none nodes is valid, all nodes's score is 0")
		scoreRes := make(extenderv1.HostPriorityList, nodeNum)
		for i := 0; i < nodeNum; i++ {
			nodeName := nodeNames[i]
			scoreRes[i] = extenderv1.HostPriority{
				Host:  nodeName,
				Score: model.MinNodeScore,
			}
		}
	}

	scoreMap := algo.BNPScore(validNames, netNeed, curArr, capArr)
	scoreRes := make(extenderv1.HostPriorityList, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodeName := nodeNames[i]
		if _, ok := scoreMap[nodeName]; !ok {
			scoreRes[i] = extenderv1.HostPriority{
				Host:  nodeName,
				Score: model.MinNodeScore,
			}
		} else {
			scoreRes[i] = extenderv1.HostPriority{
				Host:  nodeName,
				Score: scoreMap[nodeName],
			}
		}
	}

	return scoreRes
}

// BNPScore 内部评分函数
// needed 单位 Kbit/s, curMap、capMap单位Kbit/s
func (algo *BalanceNetloadPriority) BNPScore(nodeNames []string, needed int64, curMap, capMap []float64) map[string]int64 {
	nodeNum := len(nodeNames)
	// 如果needed为0，则BNP算法没有意义，所有节点评分为0
	if needed == 0 {
		scoreMap := make(map[string]int64)
		for i := 0; i < nodeNum; i++ {
			nodeName := nodeNames[i]
			scoreMap[nodeName] = model.MinNodeScore
		}

		return scoreMap
	}

	if nodeNum == 0 {
		scoreArr := make(map[string]int64)
		return scoreArr
	}

	if nodeNum == 1 {
		return map[string]int64{
			nodeNames[0]: model.MaxNodeScore,
		}
	}

	// 1. 计算当前节点的负载
	curLoad := make([]float64, nodeNum)
	for i := 0; i < nodeNum; i++ {
		curLoad[i] = curMap[i] / capMap[i]
	}

	// 2. 计算pod调度到节点i的负载
	newLoad := make([]float64, nodeNum)
	for i := 0; i < nodeNum; i++ {
		newLoad[i] = (curMap[i] + float64(needed)) / capMap[i]
	}

	// 3. 计算2中所有节点的平均负载
	// newAvgLoad := make([]float64, nodeNum)
	loadDiff := make([]float64, nodeNum)
	for i := 0; i < nodeNum; i++ {
		tmp := curLoad[i]
		curLoad[i] = newLoad[i]

		// 求均值和标准差
		// newAvgLoad[i] = stat.Mean(curLoad, nil)
		loadDiff[i] = stat.Variance(curLoad, nil)

		curLoad[i] = tmp
	}

	// 4. 计算并且正则化各个节点的得分
	loadMin, loadMax := floats.Min(loadDiff), floats.Max(loadDiff)
	loadBase := loadMax - loadMin
	log.V(5).Info("loadMax: %f, loadMin: %f, loadBase: %v", loadMax, loadMin, loadBase)
	log.V(5).Info("loadDiff: %v", loadDiff)
	scoreArr := make(map[string]int64)
	for i := 0; i < nodeNum; i++ {
		nodeName := nodeNames[i]
		if loadBase != 0.0 {
			scoreArr[nodeName] = int64(model.MaxNodeScore - (model.MaxNodeScore * (loadDiff[i] - loadMin) / loadBase))
		} else {
			scoreArr[nodeName] = model.MinNodeScore
		}
	}

	log.V(3).Info("scoreArr: %v", scoreArr)
	return scoreArr
}
