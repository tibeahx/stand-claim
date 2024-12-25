package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v4"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Postgres PostgresConfig
	Bot      BotConfig
}

type PostgresConfig struct {
	DSN struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		SslMode  string `yaml:"sslmode"`
		DbName   string `yaml:"db_name"`
	}
	MaxIdleConns int  `yaml:"max_idle_conns"`
	MaxOpenConns int  `yaml:"max_open_conns"`
	UseSeed      bool `yaml:"use_seed"`
}

var Commands []telebot.Command

type BotConfig struct {
	wrappedCommands []commandWrapper `yaml:"commands"`
	Stands          []string         `yaml:"stands"`
	Token           string           `yaml:"bot_token"`
	Verbose         bool             `yaml:"verbose"`
}

// used when config.yaml has no stands property
var defaultStands = []string{
	"dev1",
	"dev2",
	"dev3",
	"dev4",
}

type commandWrapper struct {
	Text        string `yaml:"command"`
	Description string `yaml:"description"`
}

func toTeleCommand(in []commandWrapper) []telebot.Command {
	teleCommands := make([]telebot.Command, len(in))
	for i, cmd := range in {
		teleCommands[i] = telebot.Command{
			Text:        cmd.Text,
			Description: cmd.Description,
		}
	}
	return teleCommands
}

// used when config.yaml has no commands property
var defaultCommands = []telebot.Command{
	{Text: "claim", Description: "Claim a stand"},
	{Text: "release", Description: "Release currently claimed stand"},
	{Text: "status", Description: "Show current stand status"},
	{Text: "list", Description: "Show all stands"},
	{Text: "list_free", Description: "Show available stands"},
	{Text: "ping", Description: "Ping current stand owner"},
}

var config *Config

const botTokenKey = "BOT_TOKEN"

func Load(cfgPath string) error {
	var cfg *Config

	cfgFileBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load cfg due to %w", err)
	}

	if err := yaml.Unmarshal(cfgFileBytes, &cfg); err != nil {
		return fmt.Errorf("failed to parse fileBytes due to %w", err)
	}

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load env due to %w", err)
	}

	cfg.Bot.Token = os.Getenv(botTokenKey)
	err = os.Setenv(botTokenKey, cfg.Bot.Token)
	if err != nil {
		return err
	}

	if cfg.Bot.wrappedCommands == nil {
		for _, cmd := range defaultCommands {
			if cmd.Text == "" {
				continue
			}
		}
		Commands = toTeleCommand(cfg.Bot.wrappedCommands)
	}

	if cfg.Bot.Stands == nil {
		for _, stand := range defaultStands {
			if stand == "" {
				continue
			}
			cfg.Bot.Stands = append(cfg.Bot.Stands, stand)
		}
	}

	config = cfg

	return nil
}

func Get() (*Config, error) {
	cfgPath := "config/config.yaml"
	if config == nil {
		err := Load(cfgPath)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
