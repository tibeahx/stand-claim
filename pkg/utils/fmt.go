package utils

import (
	"fmt"
	"time"

	"github.com/tibeahx/claimer/pkg/entity"
)

const (
	Free     = "✅"
	Busy     = "❌"
	Computer = "🖥️"
)

func FormatStandStatus(stand entity.Stand) string {
	if !stand.Released {
		timeBusy := time.Since(stand.TimeClaimed)

		return fmt.Sprintf(
			"busy by @%s for %d hours %s",
			stand.OwnerUsername,
			int(timeBusy.Hours()),
			Busy,
		)
	}

	return fmt.Sprintf("is finally free %s", Free)
}
