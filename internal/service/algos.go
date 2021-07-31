package service

import (
	v1 "k8s.io/api/core/v1"
)

type DemoScoreAlgo struct{}

// Score Node评分算法
func (algo *DemoScoreAlgo) Score(pod *v1.Pod, Nodes *v1.NodeList) (int64, error) {
	return 0, nil
}
