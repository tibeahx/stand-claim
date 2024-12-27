package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/tibeahx/claimer/app/internal/config"
	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/app/internal/telegram"
	"github.com/tibeahx/claimer/app/internal/worker"
	"github.com/tibeahx/claimer/pkg/log"
	"gopkg.in/telebot.v4"
)

const notifierCheckInterval = 10 * time.Second // для теста пока

func main() {
	logger := log.Zap()

	cfg, err := config.Get()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("%v", cfg)

	db, err := initDb(cfg)
	if err != nil {
		logger.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	bot, err := telegram.NewBot(cfg)
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}

	repo := repo.NewRepo(db)

	logger.Info("init repo...")

	handler := telegram.NewHandler(bot, repo)

	initCommands(bot, cfg, handler)

	info := telegram.ChatInfo

	notifier := worker.NewWorker(handler, handler.Notify(info.ChatID), 100*time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go notifier.Start(ctx, notifierCheckInterval)

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
		cancel()
		bot.Tele().Stop()
		logger.Info("shutting down...")
	}()
	wg.Wait()
}

func initDb(cfg *config.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.DSN.Host,
		cfg.Postgres.DSN.Port,
		cfg.Postgres.DSN.User,
		cfg.Postgres.DSN.Password,
		cfg.Postgres.DSN.DbName,
		cfg.Postgres.DSN.SslMode,
	)
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if cfg.Postgres.UseSeed {
		fixtures, err := testfixtures.New(
			testfixtures.Database(db.DB),
			testfixtures.Dialect("postgres"),
			testfixtures.Directory("fixtures"),
			testfixtures.DangerousSkipTestDatabaseCheck(),
		)
		if err != nil {
			return nil, err
		}

		err = fixtures.Load()
		if err != nil {
			return nil, err
		}
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	db.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)

	return db, nil
}

func initCommands(
	bot *telegram.Bot,
	cfg *config.Config,
	handler *telegram.Handler,
) {
	bot.Tele().Use(telegram.ValidateCmdMiddleware)
	bot.Tele().Use(telegram.ChatInfoMiddleware(handler))

	bot.Tele().Handle(telebot.OnUserJoined, handler.Greetings)
	bot.Tele().SetCommands(config.TeleCommands)

	for command, h := range handler.Handlers() {
		bot.Tele().Handle(command, h)
	}
}
