package watcher

import (
	"context"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// podWatcher reconciles ReplicaSets
type podWatcher struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	log    logr.Logger
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &podWatcher{}

func (r *podWatcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// set up a convinient clog object so we don't have to type request over and over again
	log := r.log.WithValues("request", request)

	// Fetch the ReplicaSet from the cache
	po := &corev1.Pod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, po)
	if errors.IsNotFound(err) {
		log.Info("The pod is removed. Continuing...")
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Could not fetch Pod")
		return reconcile.Result{}, err
	}

	// Print the Pod
	log.Info("Reconciling Pod", "container name", po.Spec.Containers[0].Name)

	return reconcile.Result{}, nil
}
