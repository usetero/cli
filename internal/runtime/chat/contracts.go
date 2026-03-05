package chat

import infrachat "github.com/usetero/cli/internal/infrastructure/chat"

var _ ChatClient = (*infrachat.Client)(nil)
