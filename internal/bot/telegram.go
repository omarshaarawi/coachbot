package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/omarshaarawi/coachbot/internal/service"
)

type TelegramBot struct {
	bot     *tgbotapi.BotAPI
	handler *Handler
	chatID  int64
}

func NewTelegramBot(token string, chatID int64, fantasyService *service.FantasyService) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	handler := NewHandler(fantasyService)

	return &TelegramBot{
		bot:     bot,
		handler: handler,
		chatID:  chatID,
	}, nil
}

func (t *TelegramBot) Start(ctx context.Context) error {
	slog.Info("Authorized on account", "username", t.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				msg := t.handler.HandleCommand(update)
				if _, err := t.bot.Send(msg); err != nil {
					slog.Error("Error sending message", "error", err)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (t *TelegramBot) SendMessage(text string) error {
	if t.chatID == 0 {
		slog.Error("Chat ID not set")
		return fmt.Errorf("chat ID not set")
	}

	msg := tgbotapi.NewMessage(t.chatID, text)
	msg.ParseMode = "Markdown"
	_, err := t.bot.Send(msg)
	if err != nil {
		slog.Error("Error sending message", "error", err)
	}
	return err
}
