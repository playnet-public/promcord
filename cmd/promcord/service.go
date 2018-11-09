package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/playnet-public/promcord/pkg/service"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

const (
	appName = "Promcord"
	appKey  = "promcord"
)

// Spec for the service
type Spec struct {
	service.BaseSpec

	Addr  string `envconfig:"metrics" required:"true" help:"metrics port"`
	Token string `envconfig:"discord_token" required:"true" help:"discord bot token"`
}

var (
	// MsgCount .
	MsgCount = stats.Int64("promcord/messages/total", "Count of messages", "1")
	// MsgLength .
	MsgLength = stats.Int64("promcord/messages/length", "Length of messages", "1")
	// MsgWordCount .
	MsgWordCount = stats.Int64("promcord/message/word/count", "Count of words in messages", "1")
	// MemberCount .
	MemberCount = stats.Int64("promcord/member/count", "Count of members", "1")
)

var (
	// Guild ID of the recorded metric
	Guild, _ = tag.NewKey("guild")
	// Channel ID of the recorded metric
	Channel, _ = tag.NewKey("channel")
	// User ID of the recorded metric
	User, _ = tag.NewKey("user")
)

var (
	// MsgCountView .
	MsgCountView = &view.View{
		Name:        "msg/count",
		Measure:     MsgCount,
		Description: "The number of messages sent",
		TagKeys:     []tag.Key{Guild, Channel, User},
		Aggregation: view.Count(),
	}
	// MsgLengthView .
	MsgLengthView = &view.View{
		Name:        "msg/length",
		Measure:     MsgLength,
		Description: "The length of messages sent",
		TagKeys:     []tag.Key{Guild, Channel, User},
		Aggregation: view.LastValue(),
	}
	// MsgWordCountView .
	MsgWordCountView = &view.View{
		Name:        "msg/word/count",
		Measure:     MsgWordCount,
		Description: "The number of words sent in messages",
		TagKeys:     []tag.Key{Guild, Channel, User},
		Aggregation: view.LastValue(),
	}
	// MemberCountView .
	MemberCountView = &view.View{
		Name:        "member/count",
		Measure:     MemberCount,
		Description: "The number of members",
		TagKeys:     []tag.Key{Guild},
		Aggregation: view.LastValue(),
	}
)

func main() {
	var svc Spec
	ctx := service.Init(appKey, appName, &svc)
	defer service.Defer(ctx)

	log.From(ctx).Info("creating prometheus exporter")
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.From(ctx).Fatal("creating prometheus exporter", zap.Error(err))
	}
	view.RegisterExporter(exporter)

	log.From(ctx).Info("registering views")
	if err := view.Register(MsgCountView); err != nil {
		log.From(ctx).Fatal("registering views", zap.Error(err))
	}
	if err := view.Register(MsgLengthView); err != nil {
		log.From(ctx).Fatal("registering views", zap.Error(err))
	}
	if err := view.Register(MsgWordCountView); err != nil {
		log.From(ctx).Fatal("registering views", zap.Error(err))
	}
	if err := view.Register(MemberCountView); err != nil {
		log.From(ctx).Fatal("registering views", zap.Error(err))
	}
	view.SetReportingPeriod(1 * time.Second)

	log.From(ctx).Info("creating discord client")
	discord, err := discordgo.New("Bot " + svc.Token)
	if err != nil {
		log.From(ctx).Fatal("creating discord client", zap.Error(err))
	}

	discord.AddHandler(messageCreateHandler(ctx))
	discord.AddHandler(memberAddHandler(ctx))
	discord.AddHandler(memberRemoveHandler(ctx))

	if err := discord.Open(); err != nil {
		log.From(ctx).Fatal("opening discord connection")
	}
	defer discord.Close()

	router := chi.NewRouter()
	router.Get("/metrics", exporter.ServeHTTP)

	var srv = http.Server{
		Addr:    svc.Addr,
		Handler: router,
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		log.From(ctx).Info("shutting down server")
		err = srv.Shutdown(ctx)
		if err != nil {
			log.From(ctx).Fatal("shutting down server", zap.Error(err))
		}
	}()

	log.From(ctx).Info("serving metrics", zap.String("addr", svc.Addr))
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.From(ctx).Fatal("serving metrics", zap.String("addr", svc.Addr), zap.Error(err))
	}

	log.From(ctx).Info("finished")
}

// handler will get called on every message and is responsible for updating the respective metrics
func messageCreateHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx := log.WithFields(ctx,
			zap.String("author", m.Author.ID),
			zap.String("channel", m.ChannelID),
			zap.String("message", m.ID),
		)

		if m.Author.ID == s.State.User.ID {
			return
		}

		var guildID string
		c, err := s.Channel(m.ChannelID)
		if err != nil {
			log.From(ctx).Error("fetching channel", zap.Error(err))
			guildID = "error"
		}

		guildID = c.GuildID
		ctx = log.WithFields(ctx,
			zap.String("guild", guildID),
		)

		ctx, err = tag.New(ctx,
			tag.Insert(Guild, guildID),
			tag.Insert(Channel, m.ChannelID),
			tag.Insert(User, m.Author.ID),
		)
		if err != nil {
			log.From(ctx).Error("adding tags", zap.Error(err))
		}

		log.From(ctx).Debug("recording metric")
		stats.Record(ctx, MsgCount.M(int64(1)))
		stats.Record(ctx, MsgLength.M(int64(len(m.Content))))
		stats.Record(ctx, MsgWordCount.M(int64(len(strings.Fields(m.Content)))))
	}
}

// handler will get called when a member enters a guild and is responsible for updating the respective metrics
func memberAddHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		ctx := log.WithFields(ctx,
			zap.String("member", m.User.ID),
			zap.String("guild", m.GuildID),
		)

		var memberCount int
		g, err := s.Guild(m.GuildID)
		if err != nil {
			log.From(ctx).Error("fetching guild", zap.Error(err))
			return
		}

		memberCount = g.MemberCount
		ctx = log.WithFields(ctx,
			zap.Int("member_count", memberCount),
		)

		ctx, err = tag.New(ctx,
			tag.Insert(Guild, m.GuildID),
		)
		if err != nil {
			log.From(ctx).Error("adding tags", zap.Error(err))
		}

		log.From(ctx).Debug("recording metric")
		stats.Record(ctx, MemberCount.M(int64(memberCount)))
	}
}

// handler will get called when a member leaves a guild and is responsible for updating the respective metrics
func memberRemoveHandler(ctx context.Context) func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		ctx := log.WithFields(ctx,
			zap.String("member", m.User.ID),
			zap.String("guild", m.GuildID),
		)

		var memberCount int
		g, err := s.Guild(m.GuildID)
		if err != nil {
			log.From(ctx).Error("fetching guild", zap.Error(err))
			return
		}

		memberCount = g.MemberCount
		ctx = log.WithFields(ctx,
			zap.Int("member_count", memberCount),
		)

		ctx, err = tag.New(ctx,
			tag.Insert(Guild, m.GuildID),
		)
		if err != nil {
			log.From(ctx).Error("adding tags", zap.Error(err))
		}

		log.From(ctx).Debug("recording metric")
		stats.Record(ctx, MemberCount.M(int64(memberCount)))
	}
}
