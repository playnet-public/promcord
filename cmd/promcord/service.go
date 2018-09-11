package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
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
	Guild string `envconfig:"discord_guild" required:"true" help:"discord guild id"`
}

var (
	// MsgCount .
	MsgCount = stats.Int64("msg/count", "Count of messages", "1")
)

var (
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
		TagKeys:     []tag.Key{Channel, User},
		Aggregation: view.Count(),
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
	view.SetReportingPeriod(1 * time.Second)

	log.From(ctx).Info("creating discord client")
	discord, err := discordgo.New("Bot " + svc.Token)
	if err != nil {
		log.From(ctx).Fatal("creating discord client", zap.Error(err))
	}

	ctx = log.WithFields(ctx, zap.String("guild", svc.Guild))

	discord.AddHandler(handler(ctx))

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
func handler(ctx context.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		ctx = log.WithFields(ctx,
			zap.String("author", m.Author.ID),
			zap.String("channel", m.ChannelID),
			zap.String("message", m.ID),
		)

		if m.Author.ID == s.State.User.ID {
			return
		}

		ctx, err := tag.New(ctx,
			tag.Insert(Channel, m.ChannelID),
			tag.Insert(User, m.Author.ID),
		)
		if err != nil {
			log.From(ctx).Error("adding tags", zap.Error(err))
		}

		log.From(ctx).Debug("recording metric", zap.String("payload", m.Message.Content))
		stats.Record(ctx, MsgCount.M(int64(1)))
	}
}
