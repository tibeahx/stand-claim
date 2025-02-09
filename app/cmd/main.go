package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/tibeahx/claimer/app/internal/config"
	gitlabwrapper "github.com/tibeahx/claimer/app/internal/gitlab"
	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/app/internal/telegram"
	"github.com/tibeahx/claimer/app/internal/workers"
	"github.com/tibeahx/claimer/pkg/log"
	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

const notifierCheckInterval = 5 * time.Hour

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

	bot, err := telegram.NewBot(cfg)
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}

	repo := repo.NewRepo(db)

	logger.Info("init repo...")

	gitlabClient, err := gitlabwrapper.NewGitlabClientWrapper(cfg,
		gitlabwrapper.WithGroupID(cfg.Gitlab.GroupID),
		gitlabwrapper.WithProjectID(cfg.Gitlab.ProjectID),
	)
	if err != nil {
		logger.Fatalf("failed to create gitlab client: %v", err)
	}

	logger.Info("init gitlab client...")

	handler := telegram.NewHandler(bot, repo, gitlabClient)

	initHandlers(bot, cfg, handler)

	logger.Info("init cmd handlers...")

	info := telegram.ChatInfo

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier := workers.NewNotifier(
		handler,
		handler.Notify(info.ChatID),
		100*time.Hour,
	)

	go notifier.Start(ctx, notifierCheckInterval)

	logger.Info("init notifier...")

	bot.Tele().Start()

	logger.Info("bot started...")

	c := make(chan os.Signal, 1)
	defer close(c)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		notifier.Stop()
		cancel()
		bot.Tele().Stop()
		db.Close()
	}()
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

		time.Sleep(time.Second)

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

func initHandlers(
	bot *telegram.Bot,
	cfg *config.Config,
	handler *telegram.Handler,
) {
	bot.Tele().Use(telegram.ChatInfoMiddleware)
	bot.Tele().Use(middleware.Recover())

	bot.Tele().SetCommands(config.TeleCommands)

	bot.Tele().Handle(
		telebot.OnUserJoined,
		handler.Greetings,
		telegram.UserJoinedMiddleware(handler),
	)

	bot.Tele().Handle(
		telebot.OnUserLeft,
		handler.Stub,
		telegram.UserLeftMiddleware(handler),
	)

	bot.Tele().Handle(telebot.OnCallback, handler.HandleCallbacks)

	for command, h := range handler.CallbackHandlers() {
		bot.Tele().Handle(command, h)
	}

	for command, h := range handler.CommandHandlers() {
		bot.Tele().Handle(command, h)
	}
}
