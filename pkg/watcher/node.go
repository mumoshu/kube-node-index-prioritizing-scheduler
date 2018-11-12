package watcher

import (
	"context"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// nodeWatcher watches nodes
type nodeWatcher struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	log    logr.Logger
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &nodeWatcher{}

func (r *nodeWatcher) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// set up a convinient clog object so we don't have to type request over and over again
	log := r.log.WithValues("request", request)

	// Fetch the ReplicaSet from the cache
	no := &corev1.Node{}
	err := r.client.Get(context.TODO(), request.NamespacedName, no)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find Node")
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Could not fetch Node")
		return reconcile.Result{}, err
	}

	// Print the Pod
	log.Info("Reconciling Node", "node name", no.ObjectMeta.Name)

	return reconcile.Result{}, nil
}
