package utils

import (
	"testing"

	"liang/internal/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

func TestGenPrioritizeJSON(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				model.ResourceNetloadKey: "1024",
			},
		},
	}
	args := extenderv1.ExtenderArgs{
		Pod:       &pod,
		NodeNames: &[]string{"node-small", "node-medium", "node-large"},
	}

	GenPrioritizeJSON(&args)
}
