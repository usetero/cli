package upload

import (
	"context"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain"
)

// MutationDeps groups GraphQL mutation dependencies needed by the uploader.
type MutationDeps struct {
	Conversations graphql.Conversations
	Messages      graphql.Messages
	Services      graphql.Services
	Policies      graphql.Policies
}

// Local ports used by upload handlers.
type messageMutations interface {
	CreateMessage(ctx context.Context, message *domain.Message) error
}
