package metrics

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// MsgCountStat .
var MsgCountStat = stats.Int64("promcord/messages/total", "Count of messages", "1")

// MsgCountView .
var MsgCountView = &view.View{
	Name:        "msg/count",
	Measure:     MsgCountStat,
	Description: "The number of messages sent",
	TagKeys:     []tag.Key{Guild, Channel, User},
	Aggregation: view.Count(),
}

// MsgCount measures the count of messages tagged with guild, channel und user ids
type MsgCount struct {
	baseMetric
	msgBase
}

// Register the metric
func (m *MsgCount) Register(ctx context.Context) error {
	return m.register(ctx, MsgCountView)
}

// Record the metric
func (m *MsgCount) Record(ctx context.Context, msg *MsgMetadata) {
	ctx, err := m.msgTags(ctx, msg)
	if err != nil {
		return
	}

	stats.Record(ctx, MsgCountStat.M(int64(1)))
}
