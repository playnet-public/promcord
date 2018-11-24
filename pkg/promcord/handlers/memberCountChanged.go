package handlers

import (
	"context"

	"github.com/playnet-public/promcord/pkg/promcord/metrics"

	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.uber.org/zap"
)

// MemberCountChanged handles all member join and leave events
type MemberCountChanged struct {
	baseHandler
	Metric *metrics.MemberCount
}

// Register the metric with OpenCensus and Discord
func (m *MemberCountChanged) Register(ctx context.Context, discord *discordgo.Session) error {
	ctx = log.WithFields(ctx, zap.String("handler", "MemberCountChanged"))

	m.Metric = &metrics.MemberCount{}

	if err := m.register(ctx, m.Metric); err != nil {
		return err
	}

	discord.AddHandler(m.BuildJoin(ctx))
	discord.AddHandler(m.BuildLeave(ctx))
	discord.AddHandler(m.BuildCreate(ctx))

	return nil
}

// BuildJoin function builder
func (m *MemberCountChanged) BuildJoin(ctx context.Context) interface{} {
	return func(s *discordgo.Session, msg *discordgo.GuildMemberAdd) {
		ctx := log.WithFields(ctx,
			zap.String("member", msg.User.ID),
			zap.String("guild", msg.GuildID),
		)

		g, err := s.Guild(msg.GuildID)
		if err != nil {
			log.From(ctx).Error("fetching guild", zap.Error(err))
			return
		}

		ctx = log.WithFields(ctx,
			zap.Int("members", g.MemberCount),
		)

		log.From(ctx).Debug("recording metrics")
		m.Metric.Record(ctx, msg.GuildID, g.MemberCount)
	}
}

// BuildLeave function builder
func (m *MemberCountChanged) BuildLeave(ctx context.Context) interface{} {
	return func(s *discordgo.Session, msg *discordgo.GuildMemberRemove) {
		ctx := log.WithFields(ctx,
			zap.String("member", msg.User.ID),
			zap.String("guild", msg.GuildID),
		)

		g, err := s.Guild(msg.GuildID)
		if err != nil {
			log.From(ctx).Error("fetching guild", zap.Error(err))
			return
		}

		ctx = log.WithFields(ctx,
			zap.Int("members", g.MemberCount),
		)

		log.From(ctx).Debug("recording metrics")
		m.Metric.Record(ctx, msg.GuildID, g.MemberCount)
	}
}

// BuildCreate function builder
func (m *MemberCountChanged) BuildCreate(ctx context.Context) interface{} {
	return func(s *discordgo.Session, event *discordgo.GuildCreate) {
		ctx := log.WithFields(ctx,
			zap.String("guild", event.ID),
		)
		g, err := s.Guild(event.ID)
		if err != nil {
			log.From(ctx).Error("fetching guild", zap.Error(err))
			return
		}
		ctx = log.WithFields(ctx,
			zap.Int("members", g.MemberCount),
		)

		log.From(ctx).Debug("recording metrics")
		m.Metric.Record(ctx, event.ID, g.MemberCount)
	}
}
