package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/tibeahx/claimer/app/internal/telegram"
	"github.com/tibeahx/claimer/pkg/log"
)

type Notifier struct {
	handler                 *telegram.Handler
	fn                      func(chatID int64, users ...string) error
	standOwnershipThreshold time.Duration
	stopCh                  chan struct{}
}

func NewNotifier(
	handler *telegram.Handler,
	notifyFn func(chatID int64, users ...string) error,
	standOwnershipThreshold time.Duration,
) *Notifier {
	return &Notifier{
		handler:                 handler,
		fn:                      notifyFn,
		standOwnershipThreshold: standOwnershipThreshold,
		stopCh:                  make(chan struct{}, 1),
	}
}

func (w *Notifier) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.WithSource(log.Zap().Desugar(), "notifier").Info("shut down")
			return
		case <-w.stopCh:
			log.WithSource(log.Zap().Desugar(), "notifier").Info("received stop signal")
			return
		case <-ticker.C:
			if err := w.execNotify(); err != nil {
				log.WithSource(log.Zap().Desugar(), "notifier").
					Sugar().
					Errorf("checkStands failed in worker due to %w", err)
				continue
			}
		}
	}
}

func (w *Notifier) execNotify() error {
	stands, err := w.handler.Repo().Stands()
	if err != nil {
		return fmt.Errorf("failed to get stands: %w", err)
	}

	usersToNotify := make([]string, 0)

	for _, stand := range stands {
		if !stand.Released && stand.OwnerUsername.String != "" {
			if time.Since(stand.TimeClaimed.Time) >= w.standOwnershipThreshold {
				usersToNotify = append(usersToNotify, stand.OwnerUsername.String)
			}
		}
	}

	if len(usersToNotify) > 0 {
		if err := w.fn(telegram.ChatInfo.ChatID, usersToNotify...); err != nil {
			return fmt.Errorf("failed to notify users: %w", err)
		}
	}

	return nil
}

func (w *Notifier) Stop() {
	w.stopCh <- struct{}{}
	close(w.stopCh)
	<-w.stopCh
}
