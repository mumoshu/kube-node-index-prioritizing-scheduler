package main

import (
	"log"
	"net/http"
	"os"
	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/extender"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/cologutil"
)

var (
	TruePredicate = extender.Predicate{
		Name: "always_true",
		Func: func(pod v1.Pod, node v1.Node) (bool, error) {
			return true, nil
		},
	}

	ZeroPriority = extender.Prioritize{
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

func main() {
	cologutil.Init(os.Getenv("LOG_LEVEL"))

	predicates := []extender.Predicate{TruePredicate}
	priorities := []extender.Prioritize{ZeroPriority}

	handler := extender.NewHandler(predicates, priorities)

	log.Print("info: server starting on the port :80")
	if err := http.ListenAndServe(":80", handler); err != nil {
		log.Fatal(err)
	}
}
