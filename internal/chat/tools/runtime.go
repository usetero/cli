package tools

import (
	"context"
	"time"
)

const toolDBTimeout = 3 * time.Second

func withToolTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), toolDBTimeout)
}
