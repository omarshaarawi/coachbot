package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/omarshaarawi/coachbot/internal/api/espn"
	"github.com/omarshaarawi/coachbot/internal/api/fantasy"
	"github.com/omarshaarawi/coachbot/internal/bot"
	"github.com/omarshaarawi/coachbot/internal/config"
	"github.com/omarshaarawi/coachbot/internal/repository/memory"
	"github.com/omarshaarawi/coachbot/internal/scheduler"
	"github.com/omarshaarawi/coachbot/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error running application", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "error", err)
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	espnClient := espn.NewClient(cfg.ESPNAPI)
	espnAPI := espn.NewAPI(espnClient)
	fantasyAPI := fantasy.NewAPI(espnAPI)

	repo := memory.NewRepository()
	fantasyService := service.NewFantasyService(fantasyAPI, repo)

	telegramBot, err := bot.NewTelegramBot(cfg.TelegramBot.Token, cfg.TelegramBot.ChatID, fantasyService)
	if err != nil {
		return err
	}

	sched, err := scheduler.NewScheduler(fantasyService, telegramBot.SendMessage)
	if err != nil {
		return err
	}

	if err := sched.Start(); err != nil {
		return err
	}
	defer func() {
		err := sched.Stop()
		if err != nil {
			slog.Error("Error stopping scheduler", "error", err)
		}
	}()

	http.HandleFunc("/", healthCheckHandler)

	go func() {
		if err := http.ListenAndServe(":80", nil); err != nil {
			slog.Error("Error starting HTTP server", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := telegramBot.Start(ctx); err != nil {
			slog.Error("Error running telegram bot", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")

	return nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
