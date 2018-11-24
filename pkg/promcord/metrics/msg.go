package metrics

import (
	"context"

	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

// MsgMetadata for easily passing on common data about a message event
type MsgMetadata struct {
	Guild   string
	Channel string
	User    string
}

type msgBase struct{}

func (m msgBase) msgTags(ctx context.Context, msg *MsgMetadata) (context.Context, error) {

	ctx, err := tag.New(ctx,
		tag.Insert(Guild, msg.Guild),
		tag.Insert(Channel, msg.Channel),
		tag.Insert(User, msg.User),
	)
	if err != nil {
		log.From(ctx).Error("adding tags", zap.Error(err))
		return ctx, err
	}

	return ctx, nil
}
