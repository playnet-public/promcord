package main

import (
	"github.com/playnet-public/promcord/pkg/promcord"
	"github.com/playnet-public/promcord/pkg/promcord/handlers"
	"github.com/playnet-public/promcord/pkg/service"

	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

const (
	appName = "Promcord"
	appKey  = "promcord"
)

// Spec for the service
type Spec struct {
	service.BaseSpec

	Addr  string `envconfig:"metrics" required:"true" help:"metrics port"`
	Token string `envconfig:"discord_token" required:"true" help:"discord bot token"`
}

func main() {
	var svc Spec
	ctx := service.Init(appKey, appName, &svc)
	defer service.Defer(ctx)

	srv, err := promcord.New(ctx, svc.Token, svc.Addr)
	if err != nil {
		log.From(ctx).Fatal("preparing server", zap.String("addr", svc.Addr), zap.Error(err))
	}

	handlers := []promcord.Handler{
		&handlers.MessageCreated{},
		&handlers.MemberCountChanged{},
	}

	err = srv.Register(ctx, handlers...)
	if err != nil {
		log.From(ctx).Fatal("registering handlers", zap.Error(err))
	}

	err = srv.Start(ctx)
	if err != nil {
		log.From(ctx).Fatal("running server", zap.Error(err))
	}

	log.From(ctx).Info("finished")
}
