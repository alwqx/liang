package service

import (
	"fmt"
	"math"
	"strconv"

	"liang/internal/model"
	"liang/internal/utils"

	"github.com/go-kratos/kratos/pkg/log"
	"gonum.org/v1/gonum/mat"
	v1 "k8s.io/api/core/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// CMDNPriority
type CMDNPriority struct{}

// Score
func (cmdn *CMDNPriority) Score(pod *v1.Pod, nodeNames []string, netCapMap map[string]int64, cacheData map[string](map[string]int64)) (extenderv1.HostPriorityList, error) {
	emptyScore := GetDefaultScore(nodeNames)
	keys := []string{model.ResourceCPUKey, model.ResourceMemKey, model.ResourceNetIOKey, model.ResourceDiskIOKey}
	if err := ValidateCacheData(keys, cacheData); err != nil {
		return emptyScore, err
	}

	// 根据资源需求、负载等因素过滤掉一些Node
	netNeed := GetPodNetIONeed(pod)
	curNetMap := cacheData[model.ResourceNetIOKey]
	validNames, _, _ := FilterNodeByNet(nodeNames, netNeed, curNetMap, netCapMap)
	if len(validNames) == 0 {
		log.Warn("none nodes is valid, all nodes's score is 0")
		return emptyScore, nil
	}

	// 同向化指标
	// TODO: 这里存在问题，因为网卡带宽能力很大，当前NetIO很小时，返回都是0
	netUsageTmpMap := CalcNetUsage(validNames, curNetMap, netCapMap)
	netArr := GetUsageArray(model.UsageUpperLimit, validNames, netUsageTmpMap)
	netCapArr := GetNetCapArr(validNames, netCapMap)

	diskMap := cacheData[model.ResourceDiskIOKey]
	diskArr := GetUsageArray(model.UsageUpperLimit, validNames, diskMap)
	cpuMap := cacheData[model.ResourceCPUKey]
	cpuArr := GetUsageArray(model.UsageUpperLimit, validNames, cpuMap)
	memMap := cacheData[model.ResourceMemKey]
	memArr := GetUsageArray(model.UsageUpperLimit, validNames, memMap)

	// 形成矩阵，计算TOPSIS结果
	nodeNum := len(validNames)
	colArr := [][]float64{cpuArr, memArr, netArr, diskArr, netCapArr}
	row := nodeNum
	col := len(colArr)
	matrix := mat.NewDense(row, col, nil)
	for i := 0; i < col; i++ {
		matrix.SetCol(i, colArr[i])
	}
	log.V(5).Info("origin resource and node matrix is: \n%v", mat.Formatted(matrix))

	topScore, err := utils.CalcTOPSIS(matrix)
	if err != nil {
		log.Error("calc topsis error: %v", err)
		log.Error("matrix is:\n%v", mat.Formatted(matrix))
		return emptyScore, err
	}

	// 结果100分制正规化
	scoreMap := cmdn.ConvertMap(validNames, topScore)
	scoreRes := make(extenderv1.HostPriorityList, nodeNum)
	var score int64
	for i := 0; i < nodeNum; i++ {
		name := nodeNames[i]
		score = 0
		if ss, ok := scoreMap[name]; ok {
			score = int64(math.Round(ss * float64(model.MaxNodeScore)))
		}
		scoreRes[i] = extenderv1.HostPriority{
			Host:  name,
			Score: score,
		}
	}

	return scoreRes, nil
}

func (cmdn *CMDNPriority) ConvertMap(nodeNames []string, scores []float64) map[string]float64 {
	num := len(nodeNames)
	if num != len(scores) {
		panic("ConvertMap num of two arrays does not equal")
	}

	res := make(map[string]float64)
	for i := 0; i < num; i++ {
		name := nodeNames[i]
		res[name] = scores[i]
	}

	return res
}

// GetNetCapArr
func GetNetCapArr(nodeNames []string, capMap map[string]int64) []float64 {
	res := make([]float64, 0)
	for _, name := range nodeNames {
		res = append(res, float64(capMap[name]))
	}

	return res
}

// CalcNetUsage 计算网络使用率/负载
// TODO: 这里存在问题，因为网卡带宽能力很大，当前NetIO很小时，返回都是0
func CalcNetUsage(nodeNames []string, curMap, capMap map[string]int64) map[string]int64 {
	resMap := make(map[string]int64)
	for _, name := range nodeNames {
		tmp := float64(curMap[name]) * 100 / float64(capMap[name])
		resMap[name] = int64(math.Round(tmp))
	}

	return resMap
}

// GetPodNetIONeed 从Pod注解中拿到Pod请求的netIO信息
func GetPodNetIONeed(pod *v1.Pod) int64 {
	var netIO int64
	for k, v := range pod.Annotations {
		if k != model.ResourceNetIOKey {
			continue
		}
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Error("parse %s to int64 error:%v", v, err)
			return netIO
		}

		netIO = vInt * model.KbitPS
		break
	}

	return netIO
}

func FilterDiskIO(nodeNames []string, diskMap map[string]int64) []float64 {
	return nil
}

func FilterNodeByNet(nodeNames []string, needNet int64, curNetMap, capNetMap map[string]int64) (valideNames []string, curArr []float64, capArr []float64) {
	nodeNum := len(nodeNames)
	for i := 0; i < nodeNum; i++ {
		nodeName := nodeNames[i]
		curNet, ok := curNetMap[nodeName]
		if !ok {
			// log.Warn("current net info of node %s does not exist, skip", nodeName)
			continue
		}

		capNet, ok1 := capNetMap[nodeName]
		// 过滤掉不存在或者资源超出的情况
		if !ok1 {
			// log.Warn("cap net info of node %s does not exist, skip", nodeName)
			continue
		}

		if needNet+curNet > capNet {
			// log.Warn("request net %d plus cur net %d overflow net cap %d, skip",
			// 	needNet, curNet, capNet)
			continue
		}

		valideNames = append(valideNames, nodeName)
		curArr = append(curArr, float64(curNet))
		capArr = append(capArr, float64(capNet))
	}

	return
}

// GetUsageArray 返回CPU/Mem等使用率信息的指标数据
func GetUsageArray(upperLimit int64, nodeNames []string, usageMap map[string]int64) []float64 {
	res := make([]float64, 0)
	for _, name := range nodeNames {
		v, ok := usageMap[name]
		if !ok {
			log.Warn("usage value of %s does not exist, skip", name)
			v = 0
		}
		if v > upperLimit {
			v = 0
		}
		res = append(res, float64(v))
	}

	return res
}

// ValidateCacheData 根据Keys验证cacheData中数据是否存在
func ValidateCacheData(keys []string, cacheData map[string](map[string]int64)) error {
	for _, key := range keys {
		if _, ok := cacheData[key]; !ok {
			return fmt.Errorf("value with key %s of cacheData does not exit", key)
		}
	}

	return nil
}
