package promcord

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/seibert-media/events/pkg/api"
	"github.com/bwmarrin/discordgo"
	"github.com/seibert-media/golibs/log"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
)

// Server prepares necessary components and metric handlers
type Server struct {
	Discord *discordgo.Session
	HTTP    *api.Server
}

// New Server for the passed in Discord token serving metrics at addr
func New(ctx context.Context, token string, addr string) (*Server, error) {
	s := &Server{}

	log.From(ctx).Info("creating discord client")
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.From(ctx).Error("creating discord client", zap.Error(err))
		return nil, err
	}

	log.From(ctx).Info("creating prometheus exporter")
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.From(ctx).Error("creating prometheus exporter", zap.Error(err))
		return nil, err
	}
	view.RegisterExporter(exporter)
	view.SetReportingPeriod(1 * time.Second)

	srv := api.New(addr, false)
	srv.Router.Get("/metrics", exporter.ServeHTTP)
	srv.Router.Get("/healthz", Health(discord))

	s.Discord = discord
	s.HTTP = srv
	return s, nil
}

// Register handlers
func (s *Server) Register(ctx context.Context, handlers ...Handler) error {
	for _, h := range handlers {
		err := h.Register(ctx, s.Discord)
		if err != nil {
			return err
		}
	}

	return nil
}

// Start the Server
func (s *Server) Start(ctx context.Context) error {
	if err := s.Discord.Open(); err != nil {
		log.From(ctx).Error("opening discord connection", zap.Error(err))
		return err
	}
	defer s.Discord.Close()

	go s.HTTP.GracefulHandler(ctx)

	err := s.HTTP.Start(ctx)
	if err != nil {
		log.From(ctx).Error("running server", zap.Error(err))
		return err
	}

	return nil
}

// Health handler checking discord status
func Health(discord *discordgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		user, err := discord.User(discord.State.User.ID)
		if err != nil || user == nil {
			status = "error"
		}
		w.Write([]byte(fmt.Sprintf(`{"status": "%s", "error": "%v"}`, status, err)))
	}
}
