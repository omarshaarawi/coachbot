package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	TelegramBot TelegramBot
	ESPNAPI     ESPNAPI
}

type TelegramBot struct {
	Token  string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	ChatID int64  `envconfig:"CHAT_ID" required:"true"`
}

type ESPNAPI struct {
	Year     string `envconfig:"YEAR" required:"true"`
	LeagueID string `envconfig:"LEAGUE_ID" required:"true"`
	SWID     string `envconfig:"SWID" required:"true"`
	ESPNS2   string `envconfig:"ESPN_S2" required:"true"`
}

func New() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
