package service

import (
	"testing"

	"liang/internal/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func TestBalanceNetloadPriority_BNPScore(t *testing.T) {
	cases := []struct {
		Name      string
		NodeNames []string
		Needed    float64
		CurArr    []float64
		CapArr    []float64
		Expected  map[string]int64
	}{
		{
			Name:      "test 0",
			NodeNames: []string{"node-small"},
			Needed:    1024,
			CurArr:    []float64{0}, // bit/s
			CapArr:    []float64{model.MbitPS * 1},
			Expected: map[string]int64{
				"node-small": 100,
			},
		},
		{
			Name:      "test 1",
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			Needed:    1024,
			CurArr:    []float64{0, 0, 0}, // bit/s
			CapArr:    []float64{model.MbitPS * 1, model.MbitPS * 1.5, model.MbitPS * 2.5},
			Expected: map[string]int64{
				"node-small":  0,
				"node-medium": 66,
				"node-large":  100,
			},
		},
		{
			Name:      "test 2",
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			Needed:    1024,
			CurArr:    []float64{1024, 1024, 1024}, // bit/s
			CapArr:    []float64{model.MbitPS * 1, model.MbitPS * 1.5, model.MbitPS * 2.5},
			Expected: map[string]int64{
				"node-small":  0,
				"node-medium": 72,
				"node-large":  100,
			},
		},
		{
			Name:      "test 3",
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			Needed:    1024,
			CurArr:    []float64{16807.00002, 17923.2, 0.0}, // bit/s
			CapArr:    []float64{model.GbitPS * 1, model.GbitPS * 1.5, model.GbitPS * 2.5},
			Expected: map[string]int64{
				"node-small":  0,
				"node-medium": 51,
				"node-large":  100,
			},
		},
		{
			Name:      "test 4",
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			Needed:    1024,
			CurArr:    []float64{0, 1024, 1024}, // bit/s
			CapArr:    []float64{model.GbitPS * 1, model.GbitPS * 1.5, model.GbitPS * 2.5},
			Expected: map[string]int64{
				"node-small":  100,
				"node-medium": 0,
				"node-large":  75,
			},
		},
		{
			Name:      "test 5",
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			Needed:    1024,
			CurArr:    []float64{512, 4096, 2048}, // bit/s
			CapArr:    []float64{model.MbitPS, model.MbitPS, model.MbitPS},
			Expected: map[string]int64{
				"node-small":  100,
				"node-medium": 0,
				"node-large":  57,
			},
		},
	}

	BDPAlgo := BalanceNetloadPriority{}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := BDPAlgo.BNPScore(tc.NodeNames, int64(tc.Needed), tc.CurArr, tc.CapArr)
			for _, nodeName := range tc.NodeNames {
				if tc.Expected[nodeName] != res[nodeName] {
					t.Errorf("test %s error, shoud be %d, but get %d\n", tc.Name, tc.Expected[nodeName], res[nodeName])
				}
			}
		})
	}
}

func TestBalanceNetloadPriority_Score(t *testing.T) {
	cases := []struct {
		Name      string
		Pod       *v1.Pod
		NodeNames []string
		CurMap    map[string]int64
		CapMap    map[string]int64
		Expected  extenderv1.HostPriorityList
	}{
		{
			Name: "test 0",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetloadKey: "1024",
					},
				},
			},
			NodeNames: []string{"node-small"},
			CurMap: map[string]int64{
				"node-small": 0,
			},
			CapMap: map[string]int64{
				"node-small": model.KbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node-small", Score: 100},
			},
		},
		{
			Name: "test 1",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetloadKey: "1024",
					},
				},
			},
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			CurMap: map[string]int64{
				"node-small":  0,
				"node-medium": 0,
				"node-large":  0,
			},
			CapMap: map[string]int64{
				"node-small":  model.KbitPS,
				"node-medium": model.MbitPS,
				"node-large":  model.GbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node-small", Score: 0},
				extenderv1.HostPriority{Host: "node-medium", Score: 99},
				extenderv1.HostPriority{Host: "node-large", Score: 100},
			},
		},
		{
			Name: "test 2",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetloadKey: "1024",
					},
				},
			},
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			CurMap: map[string]int64{
				"node-small":  512,
				"node-medium": 4096,
				"node-large":  2048,
			},
			CapMap: map[string]int64{
				"node-small":  model.MbitPS,
				"node-medium": model.MbitPS,
				"node-large":  model.MbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node-small", Score: 100},
				extenderv1.HostPriority{Host: "node-medium", Score: 0},
				extenderv1.HostPriority{Host: "node-large", Score: 57},
			},
		},
		{
			Name: "test 3",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetloadKey: "1024",
					},
				},
			},
			NodeNames: []string{"node-small", "node-medium", "node-large"},
			CurMap: map[string]int64{
				"node-small":  0,
				"node-medium": 1024,
				"node-large":  1024,
			},
			CapMap: map[string]int64{
				"node-small":  model.GbitPS,
				"node-medium": model.GbitPS * 1.5,
				"node-large":  model.GbitPS * 2.5,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node-small", Score: 100},
				extenderv1.HostPriority{Host: "node-medium", Score: 0},
				extenderv1.HostPriority{Host: "node-large", Score: 75},
			},
		},
	}

	BDPScore := BalanceNetloadPriority{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := BDPScore.Score(tc.Pod, tc.NodeNames, tc.CurMap, tc.CapMap)
			if len(res) != len(tc.Expected) {
				t.Errorf("test %s, num of res %d and Expected %d not equal",
					tc.Name, len(res), len(tc.Expected))
			}
			for i := 0; i < len(res); i++ {
				if res[i].Host != tc.Expected[i].Host {
					t.Errorf("test %s, order of res %v and Expected %v not equal",
						tc.Name, res, tc.Expected)
				}
				if res[i].Score != tc.Expected[i].Score {
					t.Errorf("test %s, score of res %d and Expected %d not equal",
						tc.Name, res[i].Score, tc.Expected[i].Score)
				}
			}
		})
	}
}
