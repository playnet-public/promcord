package metrics

import (
	"context"

	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

// MemberCountStat .
var MemberCountStat = stats.Int64("promcord/member/count", "Count of members", "1")

// MemberCountView .
var MemberCountView = &view.View{
	Name:        "member/count",
	Measure:     MemberCountStat,
	Description: "The number of members",
	TagKeys:     []tag.Key{Guild},
	Aggregation: view.LastValue(),
}

// MemberCount measures the count of messages tagged with guild, channel und user ids
type MemberCount struct {
	baseMetric
}

// Register the metric
func (m *MemberCount) Register(ctx context.Context) error {
	return m.register(ctx, MemberCountView)
}

// Record the metric
func (m *MemberCount) Record(ctx context.Context, guild string, count int) {
	ctx, err := tag.New(ctx,
		tag.Insert(Guild, guild),
	)
	if err != nil {
		log.From(ctx).Error("adding tags", zap.Error(err))
		return
	}

	stats.Record(ctx, MemberCountStat.M(int64(count)))
}
