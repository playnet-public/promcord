package metrics

import (
	"context"

	"github.com/pkg/errors"
	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	// Guild ID of the recorded metric
	Guild, _ = tag.NewKey("guild")
	// Channel ID of the recorded metric
	Channel, _ = tag.NewKey("channel")
	// User ID of the recorded metric
	User, _ = tag.NewKey("user")
)

type baseMetric struct{}

func (m baseMetric) register(ctx context.Context, v *view.View) error {
	ctx = log.WithFields(ctx, zap.String("metric", v.Name))

	log.From(ctx).Info("registering view")
	if err := view.Register(v); err != nil {
		log.From(ctx).Error("registering view", zap.Error(err))
		return errors.Wrap(err, "registering view")
	}

	return nil
}
