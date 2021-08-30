package service

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"liang/internal/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func TestCMDNPriority_Score(t *testing.T) {
	cases := []struct {
		Name      string
		Pod       *v1.Pod
		NodeNames []string
		NetCapMap map[string]int64
		CacheData map[string](map[string]int64)
		Expected  extenderv1.HostPriorityList
	}{
		{
			Name: "test 0: just one node",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "1",
					},
				},
			},
			NodeNames: []string{"node1"},
			NetCapMap: map[string]int64{
				"node1": 1024,
			},
			CacheData: map[string](map[string]int64){
				model.ResourceCPUKey: map[string]int64{
					"node1": 28,
				},
				model.ResourceMemKey: map[string]int64{
					"node1": 68,
				},
				model.ResourceDiskIOKey: map[string]int64{
					"node1": 68,
				},
				model.ResourceNetIOKey: map[string]int64{
					"node1": 20,
				},
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 100},
			},
		},
		{
			Name: "test 1: 2 nodes",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "2",
					},
				},
			},
			NodeNames: []string{"node1", "node2"},
			NetCapMap: map[string]int64{
				"node1": 1000000,
				"node2": 1500000,
			},
			CacheData: map[string](map[string]int64){
				model.ResourceCPUKey: map[string]int64{
					"node1": 25,
					"node2": 18,
				},
				model.ResourceMemKey: map[string]int64{
					"node1": 68,
					"node2": 38,
				},
				model.ResourceDiskIOKey: map[string]int64{
					"node1": 8,
					"node2": 35,
				},
				model.ResourceNetIOKey: map[string]int64{
					"node1": 18,
					"node2": 22,
				},
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 30},
				extenderv1.HostPriority{Host: "node2", Score: 70},
			},
		},
		{
			Name: "test 2: 2 nodes but 3 nodes info",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "2",
					},
				},
			},
			NodeNames: []string{"node1", "node2"},
			NetCapMap: map[string]int64{
				"node1": 1000000,
				"node2": 1500000,
				"node3": 2000000,
			},
			CacheData: map[string](map[string]int64){
				model.ResourceCPUKey: map[string]int64{
					"node1": 28,
					"node2": 48,
					"node3": 10,
				},
				model.ResourceMemKey: map[string]int64{
					"node1": 68,
					"node2": 18,
					"node3": 50,
				},
				model.ResourceDiskIOKey: map[string]int64{
					"node1": 8,
					"node2": 65,
					"node3": 23,
				},
				model.ResourceNetIOKey: map[string]int64{
					"node1": 58,
					"node2": 28,
					"node3": 38,
				},
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 41},
				extenderv1.HostPriority{Host: "node2", Score: 59},
			},
		},
		{
			Name: "test 3: 3 nodes and 3 node info",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "2",
					},
				},
			},
			NodeNames: []string{"node1", "node2", "node3"},
			NetCapMap: map[string]int64{
				"node1": 1000000,
				"node2": 1500000,
				"node3": 2000000,
			},
			CacheData: map[string](map[string]int64){
				model.ResourceCPUKey: map[string]int64{
					"node1": 28,
					"node2": 8,
					"node3": 50,
				},
				model.ResourceMemKey: map[string]int64{
					"node1": 18,
					"node2": 28,
					"node3": 5,
				},
				model.ResourceDiskIOKey: map[string]int64{
					"node1": 82,
					"node2": 51,
					"node3": 23,
				},
				model.ResourceNetIOKey: map[string]int64{
					"node1": 6,
					"node2": 26,
					"node3": 63,
				},
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 40},
				extenderv1.HostPriority{Host: "node2", Score: 67},
				extenderv1.HostPriority{Host: "node3", Score: 40},
			},
		},
		{
			Name: "test 4: 3 nodes and zero info",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "2",
					},
				},
			},
			NodeNames: []string{"node1", "node2", "node3"},
			NetCapMap: map[string]int64{
				"node1": 1000000,
				"node2": 1500000,
				"node3": 2000000,
			},
			CacheData: map[string](map[string]int64){
				model.ResourceCPUKey: map[string]int64{
					"node1": 8,
					"node2": 68,
					"node3": 20,
				},
				model.ResourceMemKey: map[string]int64{
					"node1": 16,
					"node2": 55,
					"node3": 10,
				},
				model.ResourceDiskIOKey: map[string]int64{
					"node1": 0,
					"node2": 0,
					"node3": 0,
				},
				model.ResourceNetIOKey: map[string]int64{
					"node1": 12,
					"node2": 68,
					"node3": 33,
				},
			},
			Expected: extenderv1.HostPriorityList{
				extenderv1.HostPriority{Host: "node1", Score: 10},
				extenderv1.HostPriority{Host: "node2", Score: 100},
				extenderv1.HostPriority{Host: "node3", Score: 13},
			},
		},
	}

	cmdn := CMDNPriority{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res, err := cmdn.Score(tc.Pod, tc.NodeNames, tc.NetCapMap, tc.CacheData)
			if err != nil {
				t.Errorf("test %s error: %v", tc.Name, err)
			}
			for i := range res {
				if res[i].Host != tc.Expected[i].Host {
					t.Errorf("test %s error: node name %s does not match %s",
						tc.Name, tc.Expected[i].Host, res[i].Host)
				}
				if res[i].Score != tc.Expected[i].Score {
					t.Errorf("test %s error: score %d does not match %d",
						tc.Name, tc.Expected[i].Score, res[i].Score)
				}
			}
		})
	}
}

func TestCMDNPriority_ConvertMap(t *testing.T) {
	cases := []struct {
		Name      string
		NodeNames []string
		Scores    []float64
		ExpLen    int
		Expected  map[string]float64
	}{
		{
			Name:      "test 0",
			NodeNames: []string{},
			Scores:    []float64{},
			ExpLen:    0,
			Expected:  map[string]float64{},
		},
		{
			Name:      "test 1",
			NodeNames: []string{"node1"},
			Scores:    []float64{1.0},
			ExpLen:    1,
			Expected: map[string]float64{
				"node1": 1.0,
			},
		},
		{
			Name:      "test 2",
			NodeNames: []string{"node1", "node2"},
			Scores:    []float64{1.0, 2.0},
			ExpLen:    2,
			Expected: map[string]float64{
				"node1": 1.0,
				"node2": 2.0,
			},
		},
		{
			Name:      "test 3",
			NodeNames: []string{"node1", "node2", "node3"},
			Scores:    []float64{1.0, 2.1, 3.3},
			ExpLen:    3,
			Expected: map[string]float64{
				"node1": 1.0,
				"node2": 2.1,
				"node3": 3.3,
			},
		},
	}

	cmdn := CMDNPriority{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := cmdn.ConvertMap(tc.NodeNames, tc.Scores)
			if tc.ExpLen != len(res) {
				t.Errorf("test %s error: expected len is %d, but get %d",
					tc.Name, tc.ExpLen, len(res))
			}
		})
	}
}

func TestGetPodNetIONeed(t *testing.T) {
	cases := []struct {
		Name     string
		Pod      *v1.Pod
		Expected int64
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
			Expected: 1000,
		},
		{
			Name: "test 1",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "2",
					},
				},
			},
			Expected: 2000,
		},
		{
			Name: "test 2",
			Pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						model.ResourceNetIOKey: "20",
					},
				},
			},
			Expected: 20000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := GetPodNetIONeed(tc.Pod)
			if res != tc.Expected {
				t.Errorf("test %s error: res should be %d, but get %d",
					tc.Name, tc.Expected, res)
			}
		})
	}
}

func TestFilterNodeByNet(t *testing.T) {
	cases := []struct {
		Name      string
		NodeNames []string
		NetNeed   int64
		CurNetMap map[string]int64
		CapNetMap map[string]int64
		ExpNames  []string
		ExpCurArr []float64
		ExpCapArr []float64
	}{
		{
			Name:      "test 0",
			NodeNames: []string{"node1", "node2", "node3"},
			NetNeed:   1000,
			CurNetMap: map[string]int64{
				"node1": 1000,
				"node2": 1500,
				"node3": 2000,
			},
			CapNetMap: map[string]int64{
				"node1": 3000,
				"node2": 3500,
				"node3": 2500,
			},
			ExpNames:  []string{"node1", "node2"},
			ExpCurArr: []float64{1000.0, 1500.0, 2000.0},
			ExpCapArr: []float64{3000.0, 3500.0, 2500.0},
		},
		{
			Name:      "test 1",
			NodeNames: []string{"node1", "node2", "node3"},
			NetNeed:   1000,
			CurNetMap: map[string]int64{
				"node1": 1000,
				"node2": 1500,
				"node3": 2000,
			},
			CapNetMap: map[string]int64{
				"node1": 3000,
				"node2": 3400,
				"node3": 2500,
			},
			ExpNames:  []string{"node1"},
			ExpCurArr: []float64{1000.0, 1500.0, 2000.0},
			ExpCapArr: []float64{3000.0, 3500.0, 2500.0},
		},
		{
			Name:      "test 2",
			NodeNames: []string{"node1", "node2", "node3"},
			NetNeed:   10000,
			CurNetMap: map[string]int64{
				"node1": 1000,
				"node2": 1500,
				"node3": 2000,
			},
			CapNetMap: map[string]int64{
				"node1": 3000,
				"node2": 3500,
				"node3": 2500,
			},
			ExpNames:  []string{},
			ExpCurArr: []float64{1000.0, 1500.0, 2000.0},
			ExpCapArr: []float64{3000.0, 3500.0, 2500.0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			names, _, _ := FilterNodeByNet(tc.NodeNames, tc.NetNeed, tc.CurNetMap, tc.CurNetMap)
			if reflect.DeepEqual(names, tc.ExpNames) {
				t.Errorf("test %s error: names should be %v, but get %v",
					tc.Name, tc.ExpNames, names)
			}
		})
	}
}

func TestGetUsageArray(t *testing.T) {
	cases := []struct {
		Name       string
		UpperLimit int64
		NodeNames  []string
		UsageMap   map[string]int64
		Expected   []float64
	}{
		{
			Name:       "test 0",
			NodeNames:  []string{"node1", "node2", "node3"},
			UpperLimit: 90,
			UsageMap: map[string]int64{
				"node1": 80,
				"node2": 90,
				"node3": 100,
			},
			Expected: []float64{80.0, 90.0, 0.0},
		},
		{
			Name:       "test 1",
			NodeNames:  []string{"node1", "node2", "node3"},
			UpperLimit: 60,
			UsageMap: map[string]int64{
				"node1": 40,
				"node2": 90,
				"node3": 100,
			},
			Expected: []float64{40.0, 0.0, 0.0},
		},
		{
			Name:       "test 2",
			NodeNames:  []string{"node1", "node2", "node3"},
			UpperLimit: 60,
			UsageMap: map[string]int64{
				"node1": 40,
				"node2": 50,
				"node3": 30,
			},
			Expected: []float64{40.0, 50.0, 30.0},
		},
		{
			Name:       "test 3",
			NodeNames:  []string{"node1", "node2", "node3"},
			UpperLimit: 20,
			UsageMap: map[string]int64{
				"node1": 40,
				"node2": 50,
				"node3": 30,
			},
			Expected: []float64{0.0, 0.0, 0.0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			resArr := GetUsageArray(tc.UpperLimit, tc.NodeNames, tc.UsageMap)
			for i := range resArr {
				if resArr[i]-tc.Expected[i] > 0.001 {
					t.Errorf("test %s error: %dth element of result does not match, should get %f, but get %f",
						tc.Name, i, tc.Expected[i], resArr[i])
				}
			}
		})
	}
}

type CMDNTester struct {
	Name      string
	Pod       *v1.Pod
	NodeNames []string
	NetCapMap map[string]int64
	CacheData map[string](map[string]int64)
	Expected  extenderv1.HostPriorityList

	CurArr []float64
	CapArr []float64
	CurMap map[string]int64
	CapMap map[string]int64
}

/*

{
	Name: "test 0: just one node",
	Pod: &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				model.ResourceNetIOKey: "1",
			},
		},
	},
	NodeNames: []string{"node1"},
	NetCapMap: map[string]int64{
		"node1": 1024,
	},
	CacheData: map[string](map[string]int64){
		model.ResourceCPUKey: map[string]int64{
			"node1": 28,
		},
		model.ResourceMemKey: map[string]int64{
			"node1": 68,
		},
		model.ResourceDiskIOKey: map[string]int64{
			"node1": 68,
		},
		model.ResourceNetIOKey: map[string]int64{
			"node1": 20,
		},
	},
	Expected: extenderv1.HostPriorityList{
		extenderv1.HostPriority{Host: "node1", Score: 100},
	},
},

*/

// 生成包含n个Node的测试信息
func FakeBenchCMDNData(num int) *CMDNTester {
	nodeNames := make([]string, num)
	for i := 0; i < num; i++ {
		nodeNames[i] = fmt.Sprintf("node-%d", i)
	}

	capNum := 50
	capSeed := make([]int64, capNum)
	for i := 0; i < capNum; i++ {
		base := rand.Intn(20) + 1
		capSeed[i] = int64(model.GbitPS * base / 10)
	}

	cpuMap := make(map[string]int64)
	memMap := make(map[string]int64)
	diskMap := make(map[string]int64)
	netMap := make(map[string]int64)
	capMap := make(map[string]int64)
	curArr := make([]float64, num)
	capArr := make([]float64, num)

	for i := 0; i < num; i++ {
		name := nodeNames[i]
		curV := rand.Int63n(100)
		capV := capSeed[rand.Intn(capNum)]

		curArr[i] = float64(curV)
		// curMap[name] = curV
		capArr[i] = float64(capV)
		capMap[name] = capV

		cpuMap[name] = int64(rand.Intn(100))
		memMap[name] = int64(rand.Intn(100))
		diskMap[name] = int64(rand.Intn(100))
		netMap[name] = int64(rand.Intn(100))
	}

	cacheData := map[string](map[string]int64){
		model.ResourceCPUKey:    cpuMap,
		model.ResourceMemKey:    memMap,
		model.ResourceDiskIOKey: diskMap,
		model.ResourceNetIOKey:  netMap,
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				model.ResourceNetIOKey: "1",
			},
		},
	}

	return &CMDNTester{
		Name:      "BenchCMDNTest",
		Pod:       pod,
		NodeNames: nodeNames,
		NetCapMap: capMap,
		CacheData: cacheData,

		// CurMap: curMap,
		CurArr: curArr,
		CapMap: capMap,
		CapArr: capArr,
	}
}

func benchmarkCMDNPriority_Score(n int, b *testing.B) {
	cmdn := CMDNPriority{}
	tc := FakeBenchCMDNData(n)
	// 重置定时器，去掉上面分配n个数据的操作
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmdn.Score(tc.Pod, tc.NodeNames, tc.NetCapMap, tc.CacheData)
	}
}

func BenchmarkCMDNPriority_Score10(b *testing.B) {
	benchmarkCMDNPriority_Score(10, b)
}
func BenchmarkCMDNPriority_Score100(b *testing.B) {
	benchmarkCMDNPriority_Score(100, b)
}
func BenchmarkCMDNPriority_Score1000(b *testing.B) {
	benchmarkCMDNPriority_Score(1000, b)
}
func BenchmarkCMDNPriority_Score10000(b *testing.B) {
	benchmarkCMDNPriority_Score(10000, b)
}
