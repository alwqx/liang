package utils

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/go-kratos/kratos/pkg/log"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// 生成测试JSON
func GenPrioritizeJSON(args *extenderv1.ExtenderArgs) {
	res, err := json.Marshal(args)
	if err != nil {
		log.Fatal("%v", err)
	}

	log.Info("%s", string(res))
}

// GoNumLearn 学习gonum包
func GoNumLearn() {
	data := make([]float64, 9)
	for i := 0; i < 9; i++ {
		data[i] = float64(i + 1)
	}

	m := mat.NewDense(3, 3, data)
	fmt.Println(mat.Formatted(m))
	mt := m.T()
	fmt.Println(mat.Formatted(mt))

	rv := m.RowView(0)
	fmt.Println(mat.Formatted(rv))
	cv := m.ColView(0)
	fmt.Println(mat.Formatted(cv))

	dv := m.DiagView()
	fmt.Println(mat.Formatted(dv))

	var names []string
	names = append(names, "lwq")
	names = append(names, "wr")
	fmt.Println(names)
}

// CalcTOPSIS 计算TOPSIS值，输入的数组默认已经同向化
// cpu: 	[1,2,3]
// mem:		[2,3,4]
// netload:	[1,2,5]
// netcap:	[2,4,5]
// filtered: [0.0, 0.0, 0.0, 0.0]
// 如果存在列全部为0，则默认填充1
func CalcTOPSIS(matrix *mat.Dense) ([]float64, error) {
	// 矩阵是否规范检查
	if IsMatrixEmpty(matrix) {
		return nil, fmt.Errorf("empty matrix")
	}
	row := matrix.RawMatrix().Rows
	col := matrix.RawMatrix().Cols
	if row == 1 {
		return []float64{1.0}, nil
	}

	// 检查是否存在负数
	if err := CheckMatrixNegative(matrix); err != nil {
		return nil, err
	}
	// 如果某列为全部为0，则填充1
	ResetZeroCol(matrix, 1.0)

	// 1. 按照矩阵列正规化
	maxMinArr := make([][]float64, col)
	maxMinMatrix := mat.NewDense(col, 2, nil)
	for i := 0; i < col; i++ {
		colArr := GetDenseCol(matrix, i)
		normArr := NormArray(colArr)
		// 得到max/min
		maxMin := make([]float64, 2)
		maxMin[0] = floats.Max(normArr)
		maxMin[1] = floats.Min(normArr)
		maxMinArr[i] = maxMin
		maxMinMatrix.SetRow(i, maxMin)
		matrix.SetCol(i, normArr)
	}

	// 2. 计算每个维度的距离
	// 要考虑某一个维度是不是全部是0
	resMax := make([]float64, row)
	for i := 0; i < row; i++ {
		rowArr := matrix.RawRowView(i)
		maxSum, minSum := 0.0, 0.0
		for j := 0; j < col; j++ {
			maxSum += math.Pow(rowArr[j]-maxMinArr[j][0], 2)
			minSum += math.Pow(rowArr[j]-maxMinArr[j][1], 2)
		}
		dmax := math.Sqrt(maxSum)
		dmin := math.Sqrt(minSum)

		resMax[i] = dmin / (dmin + dmax)
	}

	return resMax, nil
}

// ResetZeroCol 如果某列全部为0，填充为给定值
func ResetZeroCol(m *mat.Dense, dist float64) {
	/*
			cpu		mem		disk	net		netCap
		n1	0		3		2		3		2
		n2	0		2		0		0		3
		n3	0		9		34		2		12
	*/
	row := m.RawMatrix().Rows
	col := m.RawMatrix().Cols
	distArr := make([]float64, row)
	for i := range distArr {
		distArr[i] = dist
	}
	for i := 0; i < col; i++ {
		colArr := GetDenseCol(m, i)
		if IsZeroArray(colArr) {
			m.SetCol(i, distArr)
		}
	}
}

// IsZeroArray 判断数组是否全部为0，
// true全部为0
func IsZeroArray(arr []float64) bool {
	allZeroFlag := true
	for j := range arr {
		if allZeroFlag && arr[j] != 0.0 {
			allZeroFlag = false
		}
	}

	return allZeroFlag
}

func IsMatrixEmpty(m *mat.Dense) bool {
	row := m.RawMatrix().Rows
	col := m.RawMatrix().Cols

	return row == 0 || col == 0
}

// CheckMatrixNegative 检查topsis算法输入的数组是否合法
// 某个值不能小于0
func CheckMatrixNegative(matrix *mat.Dense) error {
	var row, col int
	row = matrix.RawMatrix().Rows
	if row == 0 {
		return fmt.Errorf("empty matrix")
	}
	col = matrix.RawMatrix().Cols
	if col == 0 {
		return fmt.Errorf("empty matrix")
	}

	for i := 0; i < col; i++ {
		colArr := GetDenseCol(matrix, i)
		for j := range colArr {
			if colArr[j] < 0 {
				return fmt.Errorf("value %f of arr %v should not < 0",
					colArr[j], colArr)
			}
		}
	}

	return nil
}

// NormTOPSISArray 正规化一个数组
func NormArray(col []float64) []float64 {
	sum := 0.0
	num := len(col)
	for _, ele := range col {
		sum += math.Pow(ele, 2)
	}

	res := make([]float64, num)
	for i := range col {
		res[i] = col[i] / sum
	}

	return res
}

// GetDenseCol 获取矩阵某一列
func GetDenseCol(m *mat.Dense, c int) []float64 {
	vec := m.ColView(c)
	size := vec.Len()
	res := make([]float64, size)
	for i := range res {
		res[i] = vec.AtVec(i)
	}

	return res
}

// GetDenseCol 获取矩阵某一列
func GetDenseRow(m *mat.Dense, r int) []float64 {
	tmp := m.RawRowView(r)
	res := make([]float64, len(tmp))
	for i := range res {
		res[i] = tmp[i]
	}

	return res
}
