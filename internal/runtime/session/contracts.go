package session

import pssyncer "github.com/usetero/cli/internal/infrastructure/powersync/syncer"

var _ pssyncer.ReadinessService = (*Service)(nil)
