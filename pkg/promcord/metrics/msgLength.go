package metrics

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// MsgLengthStat .
var MsgLengthStat = stats.Int64("promcord/messages/length", "Length of messages", "1")

// MsgLengthView .
var MsgLengthView = &view.View{
	Name:        "msg/length",
	Measure:     MsgLengthStat,
	Description: "The length of messages sent",
	TagKeys:     []tag.Key{Guild, Channel, User},
	Aggregation: view.LastValue(),
}

// MsgLength measures the length of messages tagged with guild, channel und user ids
type MsgLength struct {
	baseMetric
	msgBase
}

// Register the metric
func (m *MsgLength) Register(ctx context.Context) error {
	return m.register(ctx, MsgLengthView)
}

// Record the metric
func (m *MsgLength) Record(ctx context.Context, msg *MsgMetadata, content string) {
	ctx, err := m.msgTags(ctx, msg)
	if err != nil {
		return
	}

	stats.Record(ctx, MsgLengthStat.M(int64(len(content))))
}
