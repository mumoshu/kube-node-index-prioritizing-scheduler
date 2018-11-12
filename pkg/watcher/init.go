package watcher

import (
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"github.com/pkg/errors"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/app"
)

func Init(c app.Context) error {
	// Setup a new controller to Reconcile Pods
	c.Log.Info("Setting up pod controller")
	podController, err := controller.New("pod-controller", c.Mgr, controller.Options{
		Reconciler: &podWatcher{client: c.Mgr.GetClient(), log: c.Log.WithName("pod-watcher")},
	})
	if err != nil {
		return errors.Wrap(err, "unable to set up individual controller")
	}

	// Watch Pods
	if err := podController.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return errors.Wrap(err, "unable to watch pods")
	}

	// Setup a new controller to Reconcile Pods
	c.Log.Info("Setting up pod controller")
	nodeController, err := controller.New("node-controller", c.Mgr, controller.Options{
		Reconciler: &nodeWatcher{client: c.Mgr.GetClient(), log: c.Log.WithName("node-watcher")},
	})
	if err != nil {
		return errors.Wrap(err, "unable to set up individual controller")
	}

	// Watch Nodes
	if err := nodeController.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return errors.Wrap(err, "unable to watch nodes")
	}

	return nil
}
