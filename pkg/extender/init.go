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
	appsv1 "k8s.io/api/apps/v1"
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/apimachinery/pkg/labels"
	"encoding/json"
	"strconv"
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

	decisions := map[string]string{}

	nPodsPerNode := Prioritize{
		Name: "n_pods_per_node",
		Func: func(pod v1.Pod, nodes []v1.Node) (*schedulerapi.HostPriorityList, error) {
			ctx.Log.Info("scoring pod \"%s\"", pod.Name)

			rsKey := client.ObjectKey{pod.Namespace, pod.OwnerReferences[0].Name}
			rs := appsv1.ReplicaSet{}
			err := ctx.Mgr.GetClient().Get(context.TODO(), rsKey, &rs)
			if err != nil {
				return nil, errors.Wrap(err, "Get service from kubernetes cluster error")
			}

			d := appsv1.Deployment{}
			dKey := client.ObjectKey{pod.Namespace, rs.OwnerReferences[0].Name}
			if err := ctx.Mgr.GetClient().Get(context.TODO(), dKey, &d); err != nil {
				return nil, errors.Wrap(err, "failed getting deployment")
			}

			numStr := d.ObjectMeta.Annotations["com.github.mumoshu/n-pods-per-node"]
			desiredNum := ctx.Config.DesiredPodsPerNode

			if numStr != "" {
				numFromAnnotation, err := strconv.Atoi(numStr)
				if err != nil {
					ctx.Log.Error(err, "conversion failed. falling back to the default value", "desired", desiredNum)
					//return nil, errors.Wrap(err, "failed getting desired pod count")
				} else {
					ctx.Log.Error(err, "conversion succeeded. using the specified value", "desired", numFromAnnotation)
					desiredNum = numFromAnnotation
				}
			}

			set := labels.Set(rs.Spec.Selector.MatchLabels)
			listOpts := &client.ListOptions{
				LabelSelector: set.AsSelector(),
				Namespace: pod.Namespace,
			}
			pods := v1.PodList{}
			if listErr := ctx.Mgr.GetClient().List(context.TODO(), listOpts, &pods); listErr != nil {
				return nil, errors.Wrap(err, "error listing pods")
			}
			numScheduledPods := map[string]int{}
			for _, pod := range pods.Items {
				nodeName := pod.Spec.NodeName
				n, ok := numScheduledPods[nodeName]
				if !ok {
					numScheduledPods[nodeName] = 1
				} else {
					numScheduledPods[nodeName] = n + 1
				}
			}

			str, err := json.Marshal(numScheduledPods)
			if err != nil {
				return nil, errors.Wrap(err, "failed serializing pod counting result")
			}
			ctx.Log.Info("counted pods per nodes", "result", string(str))
			var priorityList schedulerapi.HostPriorityList
			priorityList = make([]schedulerapi.HostPriority, len(nodes))
			for i, node := range nodes {
				num, ok := numScheduledPods[node.Name]
				if !ok {
					num = 0
				} else if num >= desiredNum {
					// Provide a lower priority when the desired pod count is already fulfilled
					num = -1
				}
				score := num * 1000
				if score < 0 {
					score = 0
				}
				ctx.Log.Info("calculated score", "score", score, "num", num, "max", desiredNum, "node", node.Name)
				priorityList[i] = schedulerapi.HostPriority{
					Host:  node.Name,
					Score: score,
				}
			}
			return &priorityList, nil
		},
	}

	predicates := []Predicate{TruePredicate}
	priorities := []Prioritize{ZeroPriority, nPodsPerNode}

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
