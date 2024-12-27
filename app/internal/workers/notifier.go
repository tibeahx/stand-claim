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
	notifyFn                func(chatID int64, users ...string) error
	standOwnershipThreshold time.Duration
}

func NewNotifier(
	handler *telegram.Handler,
	notifyFn func(chatID int64, users ...string) error,
	threshold time.Duration,
) *Notifier {
	return &Notifier{
		handler:                 handler,
		notifyFn:                notifyFn,
		standOwnershipThreshold: threshold,
	}
}

func (w *Notifier) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.checkStands(); err != nil {
				log.Zap().Warnf("checkStands failed in worker due to %w", err)
				continue
			}
		}
	}
}

func (w *Notifier) checkStands() error {
	stands, err := w.handler.Repo().Stands()
	if err != nil {
		return fmt.Errorf("failed to get stands: %w", err)
	}

	var usersToNotify []string

	for _, stand := range stands {
		if !stand.Released && stand.OwnerUsername != "" {
			if time.Since(stand.TimeClaimed) >= w.standOwnershipThreshold {
				usersToNotify = append(usersToNotify, stand.OwnerUsername)
			}
		}
	}

	if len(usersToNotify) > 0 {
		if err := w.notifyFn(telegram.ChatInfo.ChatID, usersToNotify...); err != nil {
			return fmt.Errorf("failed to notify users: %w", err)
		}
	}

	return nil
}
