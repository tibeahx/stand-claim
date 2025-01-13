package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
	"gopkg.in/yaml.v3"
)

var config *Config

const botTokenKey = "BOT_TOKEN"

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Bot      BotConfig      `yaml:"bot"`
}

type PostgresConfig struct {
	DSN struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		SslMode  string `yaml:"sslmode"`
		DbName   string `yaml:"db_name"`
	} `yaml:"dsn"`
	MaxIdleConns int  `yaml:"max_idle_conns"`
	MaxOpenConns int  `yaml:"max_open_conns"`
	UseSeed      bool `yaml:"use_seed"`
}

var TeleCommands []telebot.Command

type BotConfig struct {
	RawCommands map[string]string `yaml:"commands"`
	Stands      []string          `yaml:"stands"`
	Token       string            `yaml:"bot_token"`
	Verbose     bool              `yaml:"verbose"`
}

var defaultCommands = []telebot.Command{
	{Text: "/claim", Description: "Claim a stand"},
	{Text: "/release", Description: "Release currently claimed stand"},
	{Text: "/list", Description: "Show all stands"},
	{Text: "/ping", Description: "Ping current stand owner by username"},
}

func (c *Config) teleCommandFromRaw() []telebot.Command {
	result := make([]telebot.Command, 0)

	if len(c.Bot.RawCommands) == 0 {
		TeleCommands = defaultCommands
	}

	for cmd, desc := range c.Bot.RawCommands {
		TeleCommands = append(TeleCommands, telebot.Command{
			Text:        cmd,
			Description: desc,
		})
	}

	return result
}

var errEmptyToken = errors.New("bot token is empty")

func load(cfgPath string) error {
	cfgFileBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load cfg due to %w", err)
	}

	var cfg *Config

	if err := yaml.Unmarshal(cfgFileBytes, &cfg); err != nil {
		return fmt.Errorf("failed to parse fileBytes due to %w", err)
	}

	cfg.teleCommandFromRaw()

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load env due to %w", err)
	}

	cfg.Bot.Token = os.Getenv(botTokenKey)

	if cfg.Bot.Token == "" {
		return errEmptyToken
	}

	config = cfg

	return nil
}

func Get() (*Config, error) {
	cfgPath := "config/config.yaml"

	if config == nil {
		err := load(cfgPath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
