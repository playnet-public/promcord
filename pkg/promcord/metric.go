package promcord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Handler provides the basic interface for recording metrics in promcord
// One Handler may combine multiple Metrics for handling them in the same function
type Handler interface {
	Register(ctx context.Context, discord *discordgo.Session) error
}

// Metric provides the basic interface for registering and recording single metrics
type Metric interface {
	Register(ctx context.Context) error
}
