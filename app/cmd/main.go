package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/tibeahx/claimer/app/internal/config"
	"github.com/tibeahx/claimer/app/internal/repo"
	"github.com/tibeahx/claimer/app/internal/telegram"
	"github.com/tibeahx/claimer/app/internal/workers"
	"github.com/tibeahx/claimer/pkg/entity"
	"github.com/tibeahx/claimer/pkg/log"
	"gopkg.in/telebot.v4"
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
	defer db.Close()

	bot, err := telegram.NewBot(cfg)
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}

	repo := repo.NewRepo(db)

	logger.Info("init repo...")

	handler := telegram.NewHandler(bot, repo)

	initCommands(bot, cfg, handler)

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

	logger.Info("init scheduler...")

	bot.Tele().Start()

	logger.Info("bot started...")

	populateUsers(repo, info)

	logger.Info("retrived users from chat and populated users table...")

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

func initCommands(
	bot *telegram.Bot,
	cfg *config.Config,
	handler *telegram.Handler,
) {
	bot.Tele().Use(telegram.ValidateCmdMiddleware)
	bot.Tele().Use(telegram.ChatInfoMiddleware(bot))

	bot.Tele().Handle(telebot.OnUserJoined, handler.Greetings)
	bot.Tele().SetCommands(config.TeleCommands)

	bot.Tele().Handle(telebot.OnCallback, func(c telebot.Context) error {
		data := strings.Split(c.Callback().Data, ":")
		if len(data) != 2 {
			return c.Respond()
		}

		action, standName := data[0], data[1]

		c.Message().Payload = standName

		handlers := handler.CallbackHandlers()
		if h, ok := handlers["/"+action]; ok {
			err := h(c)
			if err != nil {
				return err
			}
			return c.Respond()
		}

		return c.Respond()
	})

	for command, h := range handler.CallbackHandlers() {
		bot.Tele().Handle(command, h)
	}

	for command, h := range handler.CommandHandlers() {
		bot.Tele().Handle(command, h)
	}
}

func populateUsers(repo *repo.Repo, groupInfo entity.ChatInfo) error {
	return repo.PopulateUsers(groupInfo)
}
