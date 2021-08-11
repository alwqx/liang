package model

// pod Annotiations Key constant
const (
	PodAnnotationKey  string = "Liang"
	ResourceNetIOKey  string = "LiangNetIO"
	ResourceDiskIOKey string = "LiangDiskIO"
	ResourceCPUKey    string = "LiangCPU"
	ResourceMemKey    string = "LiangMem"

	BaseBitPS = 1
	KbitPS    = BaseBitPS * 1024
	MbitPS    = KbitPS * 1024
	GbitPS    = MbitPS * 1024

	MaxNodeScore = 100
	MinNodeScore = 0

	// 网络负载类型，分为上传负载和下载负载
	NetIOTypeUp     = "up"
	NetIOTypeDown   = "down"
	DiskIOTypeWrite = "write"
	DiskIOTypeRead  = "read"
)

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type ExtendResource struct {
	NetworkLoad  map[string]int64
	NetCardSpeed map[string]int64
}
