package app

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/go-logr/logr"
)

type Config struct {
	DesiredPodsPerNode int
}

type Context struct {
	Mgr manager.Manager
	Log logr.Logger
	Config Config
}
