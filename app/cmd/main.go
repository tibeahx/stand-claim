package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/app/internal/service"
	telegram "github.com/tibeahx/claimer/app/internal/telegram/bot"
	"github.com/tibeahx/claimer/app/internal/telegram/handler"
	middleware "github.com/tibeahx/claimer/app/internal/transport"
	"github.com/tibeahx/claimer/pkg/log"
	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
)

const botTokenKey = "BOT_TOKEN"

func main() {
	logger := log.Zap()

	db, err := initDb(logger.Desugar(), filepath.Join(".", "/app/internal/repo", "conn.db"))
	if err != nil {
		logger.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	bot, err := telegram.NewBot(os.Getenv(botTokenKey), telegram.BotOptions{
		Verbose: true,
		ErrHandler: func(err error, c telebot.Context) {
			logger.Errorf("bot error: %v", err)
		},
	})
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}
	bot.SetCommands()

	repo, err := repo.NewRepo(db)
	if err != nil {
		logger.Fatal(err)
	}

	service := service.NewService(repo)

	logger.Info("init service...")

	handler := handler.NewHandler(bot, service)

	initCommands(bot, handler)

	logger.Info("init cmd handlers...")

	bot.Tele().Start()

	logger.Info("bot started...")

	var wg sync.WaitGroup

	closeCh := make(chan os.Signal, 1)
	defer close(closeCh)

	signal.Notify(closeCh, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-closeCh
		bot.Tele().Stop()
		logger.Info("shutting down...")
	}()
	wg.Wait()
}

const (
	sqlite3        = "sqlite3"
	maxIdleConns   = 5
	maxIdleTimeout = 60 * time.Second
	maxOpenConns   = 5
)

func initDb(logger *zap.Logger, path string) (*sqlx.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logger.Warn("failed to create dir for db", zap.Error(err))
	}

	db, err := sqlx.Open(sqlite3, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(maxIdleTimeout)
	db.SetMaxOpenConns(maxOpenConns)

	return db, nil
}

func initCommands(bot *telegram.Bot, handler *handler.Handler) {
	bot.Tele().Use(middleware.Middleware)

	for cmd, h := range handler.Handlers() {
		if cmd != "" {
			bot.Tele().Handle(cmd, h)
		}
	}
}
