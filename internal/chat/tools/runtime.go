package tools

import (
	"context"
	"time"

	"github.com/usetero/cli/internal/sqlite"
)

const toolDBTimeout = 3 * time.Second

func withToolTimeout() (context.Context, context.CancelFunc) {
	return sqlite.WithTimeout(context.Background(), toolDBTimeout)
}
