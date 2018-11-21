package metrics

import (
	"context"
	"strings"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// MsgWordCountStat .
var MsgWordCountStat = stats.Int64("promcord/message/word/count", "Count of words in messages", "1")

// MsgWordCountView .
var MsgWordCountView = &view.View{
	Name:        "msg/word/count",
	Measure:     MsgWordCountStat,
	Description: "The number of words sent in messages",
	TagKeys:     []tag.Key{Guild, Channel, User},
	Aggregation: view.LastValue(),
}

// MsgWordCount measures the amount of words in a message tagged with guild, channel und user ids
type MsgWordCount struct {
	baseMetric
	msgBase
}

// Register the metric
func (m *MsgWordCount) Register(ctx context.Context) error {
	return m.register(ctx, MsgWordCountView)
}

// Record the metric
func (m *MsgWordCount) Record(ctx context.Context, msg *MsgMetadata, content string) {
	ctx, err := m.msgTags(ctx, msg)
	if err != nil {
		return
	}

	stats.Record(ctx, MsgWordCountStat.M(int64(len(strings.Fields(content)))))
}
