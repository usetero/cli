package tools

import (
	"context"
	"time"
)

const toolTimeout = 3 * time.Second

func withToolTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), toolTimeout)
}
