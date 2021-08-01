package model

// pod Annotiations Key constant
const (
	PodAnnotationKey   string = "Liang"
	ResourceNetloadKey string = "LiangNetload"
	KbitPS                    = 1024
	MbitPS                    = KbitPS * 1024
	GbitPS                    = MbitPS * 1024

	MaxNodeScore = 100
	MinNodeScore = 0
)

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type ExtendResource struct {
	NetworkLoad  map[string]int64
	NetCardSpeed map[string]int64
}
