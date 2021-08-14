package utils

import (
	"testing"

	"liang/internal/model"

	"gonum.org/v1/gonum/mat"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func TestGenPrioritizeJSON(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				model.ResourceNetIOKey: "1024",
			},
		},
	}
	args := extenderv1.ExtenderArgs{
		Pod:       &pod,
		NodeNames: &[]string{"node-small", "node-medium", "node-large"},
	}

	GenPrioritizeJSON(&args)
}

func TestGoNumLearn(t *testing.T) {
	GoNumLearn()
}

func TestCalcTOPSIS(t *testing.T) {
	matrix11 := mat.NewDense(1, 1, []float64{
		1,
	})
	matrix12 := mat.NewDense(1, 2, []float64{
		1, 2,
	})
	matrix22 := mat.NewDense(2, 2, []float64{
		1, 2,
		3, 4,
	})
	matrix33 := mat.NewDense(3, 3, []float64{
		3, 2, 3,
		4, 4, 5,
		3, 5, 8,
	})
	matrix43 := mat.NewDense(4, 3, []float64{
		3, 2, 3,
		4, 4, 5,
		3, 5, 8,
		1, 9, 3,
	})
	matrix34 := mat.NewDense(3, 4, []float64{
		0.0002333, 0.0004232, 0.00094723, 0.000332654,
		0.0002453, 0.0006331, 0.00012323, 0.000223488,
		0.0009743, 0.0002321, 0.00023223, 0.000942234,
	})

	cases := []struct {
		Name     string
		Input    *mat.Dense
		Expected []float64
	}{
		{
			Name:     "test 0",
			Input:    matrix11,
			Expected: []float64{1.0},
		},
		{
			Name:     "test 1",
			Input:    matrix12,
			Expected: []float64{1.0},
		},
		{
			Name:     "test 2",
			Input:    matrix22,
			Expected: []float64{0.0, 1.0},
		},
		{
			Name:     "test 3",
			Input:    matrix33,
			Expected: []float64{0.0, 0.601379, 0.740548},
		},
		{
			Name:     "test 4",
			Input:    matrix43,
			Expected: []float64{0.422785, 0.647250, 0.644887, 0.362681},
		},
		{
			Name:     "test 5",
			Input:    matrix34,
			Expected: []float64{0.48522022011134025, 0.32893968322846917, 0.5026228244012292},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res, err := CalcTOPSIS(tc.Input)
			if err != nil {
				t.Errorf("test %s error: %v", tc.Name, err)
			}

			if len(res) != len(tc.Expected) {
				t.Errorf("test %s error: len should be %d, but get %d",
					tc.Name, len(tc.Expected), len(res))
			}

			for i := range res {
				if res[i]-tc.Expected[i] > 0.000001 {
					t.Errorf("test %s error: %dth element should equal, expected %f, but get %f",
						tc.Name, i, tc.Expected[i], res[i])
				}
			}
		})
	}

}

func TestNormArray(t *testing.T) {
	cases := []struct {
		Name     string
		Input    []float64
		Expected []float64
	}{
		{
			Name:     "test 0",
			Input:    []float64{1.0, 1.0, 1.0},
			Expected: []float64{0.5773502, 0.5773502, 0.5773502},
		},
		{
			Name:     "test 1",
			Input:    []float64{1.0, 2.0},
			Expected: []float64{0.44721359, 0.89442719},
		},
		{
			Name:     "test 2",
			Input:    []float64{3.0, 4.0},
			Expected: []float64{0.6, 0.8},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := NormArray(tc.Input)
			for i := range res {
				if res[i]-tc.Expected[i] > 0.00001 {
					t.Errorf("test %s error: %dth element does not equal, should %f, get %f",
						tc.Name, i, tc.Expected[i], res[i])
				}
			}
		})
	}
}
