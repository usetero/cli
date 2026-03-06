package chat

import (
	"context"
	"fmt"
	"sync"

	domainchat "github.com/usetero/cli/internal/domains/chat"
	chattools "github.com/usetero/cli/internal/domains/chat/tools"
	infrachat "github.com/usetero/cli/internal/infrastructure/chat"
)

type ChatClient interface {
	Stream(ctx context.Context, req infrachat.Request, onEvent func(infrachat.Event)) (infrachat.StreamResult, error)
}

type Runtime struct {
	conversations domainchat.ConversationService
	messages      domainchat.MessageService
	client        ChatClient
	tools         chattools.Toolset

	mu      sync.RWMutex
	state   State
	history []infrachat.Message
	updates chan State
	cancel  context.CancelFunc
	closed  bool
}

func New(
	conversations domainchat.ConversationService,
	messages domainchat.MessageService,
	client ChatClient,
) *Runtime {
	return NewWithTools(conversations, messages, client, chattools.Toolset{})
}

func NewWithTools(
	conversations domainchat.ConversationService,
	messages domainchat.MessageService,
	client ChatClient,
	tools chattools.Toolset,
) *Runtime {
	if conversations == nil {
		panic("chat runtime requires conversation service")
	}
	if messages == nil {
		panic("chat runtime requires message service")
	}
	if client == nil {
		panic("chat runtime requires client")
	}

	return &Runtime{
		conversations: conversations,
		messages:      messages,
		client:        client,
		tools:         tools,
		updates:       make(chan State, 32),
		state:         State{CanSend: true},
	}
}

func (r *Runtime) State() State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return cloneState(r.state)
}

func (r *Runtime) Updates() <-chan State { return r.updates }

func (r *Runtime) requireReadyToSend() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("chat runtime is closed")
	}
	if r.state.Streaming {
		return fmt.Errorf("chat stream already in progress")
	}
	return nil
}

func (r *Runtime) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	r.mu.Unlock()
	return nil
}
