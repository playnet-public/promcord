package main

import (
	"github.com/playnet-public/promcord/pkg/service"

	"github.com/seibert-media/golibs/log"
)

const (
	appName = "Promcord"
	appKey  = "promcord"
)

// Spec for the service
type Spec struct {
	service.BaseSpec
}

func main() {
	var svc Spec
	ctx := service.Init(appKey, appName, &svc)
	defer service.Defer(ctx)

	log.From(ctx).Info("finished")
}
