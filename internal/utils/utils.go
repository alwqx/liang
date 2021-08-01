package utils

import (
	"encoding/json"

	"github.com/go-kratos/kratos/pkg/log"
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
