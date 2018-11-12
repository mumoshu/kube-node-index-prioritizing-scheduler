package extender

import (
	"log"
	"net/http"
	"os"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/cologutil"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/app"
	"github.com/go-logr/logr"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	"k8s.io/api/core/v1"
	"encoding/json"
	"sort"
	"fmt"
)

type runnable struct {
	log logr.Logger
	handler *httprouter.Router

	Started bool
}

func (r runnable) Start(stop <-chan struct{}) error {
	r.log.Info("Starting extender")
	srv := startHttpServer(r.handler)
	r.Started = true

	<-stop

	r.log.Info("Stopping extender")
	if err := srv.Shutdown(nil); err != nil {
		return errors.Wrap(err, "failure/timeout shutting down the server gracefully")
	}

	return nil
}

func Init(ctx app.Context) error {
	cologutil.Init(os.Getenv("LOG_LEVEL"))

	nPodsPerNode := Prioritize{
		Name: "node_index",
		Func: func(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
			ctx.Log.Info(fmt.Sprintf("scoring pod \"%s\"", pod.Name))

			sorted := nodes

			sort.Slice(sorted, func(i, j int) bool {
				return nodes[i].Name > nodes[j].Name
			})

			scores := map[string]int{}

			var priorityList schedulerapi.HostPriorityList
			priorityList = make([]schedulerapi.HostPriority, len(nodes))
			max := 10
			for i, n := range sorted {
				scores[n.Name] = max - i
				if scores[n.Name] < 0 {
					scores[n.Name] = 0
				}

				priorityList[i] = schedulerapi.HostPriority{
					Host:  n.Name,
					Score: scores[n.Name],
				}
			}

			str, err := json.Marshal(priorityList)
			if err != nil {
				return nil, errors.Wrap(err, "failed serializing pod counting result")
			}
			ctx.Log.Info("host priority list produced", "result", string(str))

			return &priorityList, nil
		},
	}

	predicates := []Predicate{}
	priorities := []Prioritize{nPodsPerNode}

	handler := NewHandler(predicates, priorities)

	extender := runnable{
		log: ctx.Log.WithName("extender"),
		handler: handler,
	}

	if err := ctx.Mgr.Add(extender); err != nil {
		return errors.Wrap(err, "failed to add extender")
	}

	return nil
}

func startHttpServer(handler *httprouter.Router) http.Server {
	log.Print("info: server starting on the port :80")

	addr := ":80"
	srv := http.Server{Addr: addr, Handler: handler,}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	return srv
}
