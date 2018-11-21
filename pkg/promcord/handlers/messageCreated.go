package handlers

import (
	"context"

	"github.com/playnet-public/promcord/pkg/promcord/metrics"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

// MessageCreated handles all new messages and updates the respective metrics.
// Current Metrics include: MsgCount, MsgLength, MsgWordCount
type MessageCreated struct {
	baseHandler
	Metrics messageCreatedMetrics
}

type messageCreatedMetrics struct {
	MsgCount     *metrics.MsgCount
	MsgLength    *metrics.MsgLength
	MsgWordCount *metrics.MsgWordCount
}

// Register the metric with OpenCensus and Discord
func (m *MessageCreated) Register(ctx context.Context, discord *discordgo.Session) error {
	ctx = log.WithFields(ctx, zap.String("handler", "MessageCreated"))

	m.Metrics = messageCreatedMetrics{
		&metrics.MsgCount{},
		&metrics.MsgLength{},
		&metrics.MsgWordCount{},
	}

	if err := m.register(ctx, m.Metrics.MsgCount); err != nil {
		return err
	}
	if err := m.register(ctx, m.Metrics.MsgLength); err != nil {
		return err
	}
	if err := m.register(ctx, m.Metrics.MsgWordCount); err != nil {
		return err
	}

	discord.AddHandler(m.Build(ctx))

	return nil
}

// Build function builder
func (m *MessageCreated) Build(ctx context.Context) interface{} {
	return func(s *discordgo.Session, msg *discordgo.MessageCreate) {
		ctx := log.WithFields(ctx,
			zap.String("guild", msg.GuildID),
			zap.String("channel", msg.ChannelID),
			zap.String("author", msg.Author.ID),
			zap.String("message", msg.ID),
		)

		if msg.Author.ID == s.State.User.ID {
			return
		}

		meta := &metrics.MsgMetadata{
			Guild:   msg.GuildID,
			Channel: msg.ChannelID,
			User:    msg.Author.ID,
		}

		log.From(ctx).Debug("recording metrics")
		m.Metrics.MsgCount.Record(ctx, meta)
		m.Metrics.MsgLength.Record(ctx, meta, msg.Content)
		m.Metrics.MsgWordCount.Record(ctx, meta, msg.Content)
	}
}
