package main

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/app"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/watcher"
	"github.com/everpeace/k8s-scheduler-extender-example/pkg/extender"
)

var log = logf.Log.WithName("n-pods-per-node")

func main() {
	//var disableWebhookConfigInstaller bool
	//flag.BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false,
	//	"disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping")
	//
	//flag.Parse()
	logf.SetLogger(logf.ZapLogger(false))
	entryLog := log.WithName("entrypoint")

	// Setup a Manager
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	ctx := app.Context{
		Mgr: mgr,
		Log: log,
		Config: app.Config{
			DesiredPodsPerNode: 3,
		},
	}

	if err := watcher.Init(ctx); err != nil {
		log.Error(err, "%+v")
		os.Exit(1)
	}

	if err := extender.Init(ctx); err != nil {
		log.Error(err, "%+v")
		os.Exit(1)
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
