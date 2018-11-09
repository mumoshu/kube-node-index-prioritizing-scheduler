package extender

import (
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/api/core/v1"
)

var (
	TruePredicate = Predicate{
		Name: "always_true",
		Func: func(pod v1.Pod, node v1.Node) (bool, error) {
			return true, nil
		},
	}

	ZeroPriority = Prioritize{
		Name: "zero_score",
		Func: func(_ v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
			var priorityList schedulerapi.HostPriorityList
			priorityList = make([]schedulerapi.HostPriority, len(nodes))
			for i, node := range nodes {
				priorityList[i] = schedulerapi.HostPriority{
					Host:  node.Name,
					Score: 0,
				}
			}
			return &priorityList, nil
		},
	}
)
