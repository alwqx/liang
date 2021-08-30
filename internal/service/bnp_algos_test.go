package service

import (
	"fmt"
	"math/rand"
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
			NodeNames: []string{"node1"},
			Needed:    1,
			CurArr:    []float64{0}, // bit/s
			CapArr:    []float64{model.MbitPS * 1},
			Expected: map[string]int64{
				"node1": 100,
			},
		},
		{
			Name:      "test 1",
			NodeNames: []string{"node1", "node2", "node3"},
			Needed:    1,
			CurArr:    []float64{0, 0, 0}, // bit/s
			CapArr:    []float64{model.MbitPS * 1, model.MbitPS * 1.5, model.MbitPS * 2.5},
			Expected: map[string]int64{
				"node1": 0,
				"node2": 66,
				"node3": 100,
			},
		},
		{
			Name:      "test 2",
			NodeNames: []string{"node1", "node2", "node3"},
			Needed:    1,
			CurArr:    []float64{1024, 1024, 1024}, // bit/s
			CapArr:    []float64{model.MbitPS * 1, model.MbitPS * 1.5, model.MbitPS * 2.5},
			Expected: map[string]int64{
				"node1": 0,
				"node2": 76,
				"node3": 100,
			},
		},
		{
			Name:      "test 3",
			NodeNames: []string{"node1", "node2", "node3"},
			Needed:    1,
			CurArr:    []float64{16807.00002, 17923.2, 0.0}, // bit/s
			CapArr:    []float64{model.GbitPS * 1, model.GbitPS * 1.5, model.GbitPS * 2.5},
			Expected: map[string]int64{
				"node1": 0,
				"node2": 51,
				"node3": 100,
			},
		},
		{
			Name:      "test 4",
			NodeNames: []string{"node1", "node2", "node3"},
			Needed:    1,
			CurArr:    []float64{0, 1024, 1024}, // bit/s
			CapArr:    []float64{model.GbitPS * 1, model.GbitPS * 1.5, model.GbitPS * 2.5},
			Expected: map[string]int64{
				"node1": 100,
				"node2": 0,
				"node3": 33,
			},
		},
		{
			Name:      "test 5",
			NodeNames: []string{"node1", "node2", "node3", "node4"},
			Needed:    1,
			CurArr:    []float64{512, 4096, 2048, 1024}, // bit/s
			CapArr:    []float64{model.MbitPS, model.MbitPS, model.MbitPS * 2, model.MbitPS * 3},
			Expected: map[string]int64{
				"node1": 100,
				"node2": 0,
				"node3": 79,
				"node4": 83,
			},
		},
		{
			Name:      "test 6",
			NodeNames: []string{"node1", "node2", "node3", "node4", "node5", "node6"},
			Needed:    2,
			CurArr:    []float64{512, 4096, 2048, 512, 1024, 1024}, // Kbit/s
			CapArr:    []float64{model.MbitPS, model.MbitPS * 2, model.MbitPS, model.MbitPS * 5, model.GbitPS, model.MbitPS},
			Expected: map[string]int64{
				"node1": 100,
				"node2": 35,
				"node3": 0,
				"node4": 82,
				"node5": 71,
				"node6": 66,
			},
		},
		{
			Name:      "test 7",
			NodeNames: []string{"node1", "node2", "node3"},
			Needed:    2,
			CurArr:    []float64{7, 83049, 461394}, // Kbit/s
			CapArr:    []float64{model.MbitPS, model.MbitPS * 1.5, model.MbitPS * 2.5},
			Expected: map[string]int64{
				"node1": 100,
				"node2": 47,
				"node3": 0,
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
						model.ResourceNetIOKey: "1",
					},
				},
			},
			NodeNames: []string{"node1"},
			CurMap: map[string]int64{
				"node1": 0,
			},
			CapMap: map[string]int64{
				"node1": model.KbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 100},
			},
		},
		{
			Name: "test 1",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "1",
					},
				},
			},
			NodeNames: []string{"node1", "node2", "node3"},
			CurMap: map[string]int64{
				"node1": 0,
				"node2": 0,
				"node3": 0,
			},
			CapMap: map[string]int64{
				"node1": model.KbitPS,
				"node2": model.MbitPS,
				"node3": model.GbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 0},
				extenderv1.HostPriority{Host: "node2", Score: 99},
				extenderv1.HostPriority{Host: "node3", Score: 100},
			},
		},
		{
			Name: "test 2",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "1",
					},
				},
			},
			NodeNames: []string{"node1", "node2", "node3"},
			CurMap: map[string]int64{
				"node1": 512,
				"node2": 4096,
				"node3": 2048,
			},
			CapMap: map[string]int64{
				"node1": model.MbitPS,
				"node2": model.MbitPS,
				"node3": model.MbitPS,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 100},
				extenderv1.HostPriority{Host: "node2", Score: 0},
				extenderv1.HostPriority{Host: "node3", Score: 57},
			},
		},
		{
			Name: "test 3",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "1",
					},
				},
			},
			NodeNames: []string{"node1", "node2", "node3"},
			CurMap: map[string]int64{
				"node1": 0,
				"node2": 1024,
				"node3": 1024,
			},
			CapMap: map[string]int64{
				"node1": model.GbitPS,
				"node2": model.GbitPS * 1.5,
				"node3": model.GbitPS * 2.5,
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 100},
				extenderv1.HostPriority{Host: "node2", Score: 0},
				extenderv1.HostPriority{Host: "node3", Score: 73},
			},
		},
	}

	BDPScore := BalanceNetloadPriority{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res, err := BDPScore.Score(tc.Pod, tc.NodeNames, tc.CurMap, tc.CapMap)
			if err != nil {
				t.Errorf("test %s error: %v", tc.Name, err)
			}
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

// 模拟100个Node，1000个Node和10000个Node的算法性能
type BNPTester struct {
	Name      string
	Pod       *v1.Pod
	NeedNetIO int
	NodeNames []string
	CurArr    []float64
	CapArr    []float64
	CurMap    map[string]int64
	CapMap    map[string]int64
	Expected  extenderv1.HostPriorityList
}

// 生成包含n个Node的测试信息
func FakeBenchBNPData(num int) *BNPTester {
	nodeNames := make([]string, num)
	for i := 0; i < num; i++ {
		nodeNames[i] = fmt.Sprintf("node-%d", i)
	}

	var capSeed []int64
	capNum := 50
	for i := 0; i < capNum; i++ {
		base := rand.Intn(10) + 1
		capSeed = append(capSeed, int64(model.GbitPS*base/10))
	}
	curMap := make(map[string]int64)
	capMap := make(map[string]int64)
	curArr := make([]float64, num)
	capArr := make([]float64, num)
	for i := 0; i < num; i++ {
		name := nodeNames[i]
		curV := rand.Int63n(100)
		capV := capSeed[rand.Intn(capNum)]

		curArr[i] = float64(curV)
		curMap[name] = curV
		capArr[i] = float64(capV)
		capMap[name] = capV
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				model.ResourceNetIOKey: "1",
			},
		},
	}

	return &BNPTester{
		Name:      "BenchTest",
		Pod:       pod,
		NeedNetIO: 10,
		NodeNames: nodeNames,
		CurMap:    curMap,
		CurArr:    curArr,
		CapMap:    capMap,
		CapArr:    capArr,
	}
}

func benchmarkBalanceNetloadPriority_BNPScore(n int, b *testing.B) {
	BDPAlgo := BalanceNetloadPriority{}
	tc := FakeBenchBNPData(n)
	// 重置定时器，去掉上面分配n个数据的操作
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BDPAlgo.BNPScore(tc.NodeNames, int64(tc.NeedNetIO), tc.CurArr, tc.CapArr)
	}
}

func BenchmarkBalanceNetloadPriority_BNPScore10(b *testing.B) {
	benchmarkBalanceNetloadPriority_BNPScore(10, b)
}
func BenchmarkBalanceNetloadPriority_BNPScore100(b *testing.B) {
	benchmarkBalanceNetloadPriority_BNPScore(100, b)
}
func BenchmarkBalanceNetloadPriority_BNPScore1000(b *testing.B) {
	benchmarkBalanceNetloadPriority_BNPScore(1000, b)
}
func BenchmarkBalanceNetloadPriority_BNPScore10000(b *testing.B) {
	benchmarkBalanceNetloadPriority_BNPScore(10000, b)
}

func benchmarkBalanceNetloadPriority_Score(n int, b *testing.B) {
	BDPAlgo := BalanceNetloadPriority{}
	tc := FakeBenchBNPData(n)
	// 重置定时器，去掉上面分配n个数据的操作
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BDPAlgo.Score(tc.Pod, tc.NodeNames, tc.CurMap, tc.CapMap)
	}
}

func BenchmarkBalanceNetloadPriority_Score10(b *testing.B) {
	benchmarkBalanceNetloadPriority_BNPScore(10, b)
}
func BenchmarkBalanceNetloadPriority_Score100(b *testing.B) {
	benchmarkBalanceNetloadPriority_Score(100, b)
}
func BenchmarkBalanceNetloadPriority_Score1000(b *testing.B) {
	benchmarkBalanceNetloadPriority_Score(1000, b)
}
func BenchmarkBalanceNetloadPriority_Score10000(b *testing.B) {
	benchmarkBalanceNetloadPriority_Score(10000, b)
}
