package extender

import (
	"fmt"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	NoBind = Bind{
		Func: func(podName string, podNamespace string, podUID types.UID, node string) error {
			return fmt.Errorf("This extender doesn't support Bind.  Please make 'BindVerb' be empty in your ExtenderConfig.")
		},
	}

	EchoPreemption = Preemption{
		Func: func(
			_ v1.Pod,
			_ map[string]*schedulerapi.Victims,
			nodeNameToMetaVictims map[string]*schedulerapi.MetaVictims,
		) map[string]*schedulerapi.MetaVictims {
			return nodeNameToMetaVictims
		},
	}
)
