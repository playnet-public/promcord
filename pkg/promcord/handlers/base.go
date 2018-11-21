package handlers

import (
	"context"

	"github.com/playnet-public/promcord/pkg/promcord"

	"github.com/pkg/errors"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

type baseHandler struct{}

func (h baseHandler) register(ctx context.Context, metrics ...promcord.Metric) error {
	for _, m := range metrics {
		err := m.Register(ctx)
		if err != nil {
			log.From(ctx).Error("registering metric", zap.Error(err))
			return errors.Wrap(err, "registering metric")
		}
	}

	return nil
}
