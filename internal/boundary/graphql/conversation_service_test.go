package graphql_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/boundary/graphql/apitest"
	"github.com/usetero/cli/internal/boundary/graphql/gen"
	"github.com/usetero/cli/internal/log/logtest"
)

func TestConversationService_Create_UsesCurrentContract(t *testing.T) {
	t.Parallel()

	var capturedInput gen.CreateConversationInput
	mockClient := &apitest.MockClient{
		CreateConversationFunc: func(ctx context.Context, input gen.CreateConversationInput) (*gen.CreateConversationResponse, error) {
			capturedInput = input
			return &gen.CreateConversationResponse{
				CreateConversation: gen.CreateConversationCreateConversation{
					Id: "conv-1",
				},
			}, nil
		},
	}

	svc := graphql.NewConversationService(mockClient, logtest.NewScope(t))
	_, err := svc.Create(context.Background(), graphql.CreateConversationInput{
		ID:        uuid.New(),
		AccountID: "acc-123",
		Title:     "Test conversation",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if capturedInput.AccountID == nil || *capturedInput.AccountID != "acc-123" {
		t.Fatalf("AccountID = %v, want acc-123", capturedInput.AccountID)
	}
	if capturedInput.WorkspaceID != nil {
		t.Fatalf("WorkspaceID = %v, want nil", *capturedInput.WorkspaceID)
	}
}

func TestConversationService_Update_UsesCurrentContract(t *testing.T) {
	t.Parallel()

	var capturedInput gen.UpdateConversationInput
	mockClient := &apitest.MockClient{
		UpdateConversationFunc: func(ctx context.Context, id string, input gen.UpdateConversationInput) (*gen.UpdateConversationResponse, error) {
			capturedInput = input
			return &gen.UpdateConversationResponse{
				UpdateConversation: gen.UpdateConversationUpdateConversation{
					Id:    id,
					Title: input.Title,
				},
			}, nil
		},
	}

	svc := graphql.NewConversationService(mockClient, logtest.NewScope(t))
	title := "Renamed"
	_, err := svc.Update(context.Background(), "conv-1", graphql.UpdateConversationInput{Title: &title})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if capturedInput.Title == nil || *capturedInput.Title != title {
		t.Fatalf("Title = %v, want %q", capturedInput.Title, title)
	}
	if capturedInput.ClearTitle != nil {
		t.Fatalf("ClearTitle = %v, want nil", *capturedInput.ClearTitle)
	}
}

func TestConversationService_Update_ClearsTitleWithCurrentContract(t *testing.T) {
	t.Parallel()

	var capturedInput gen.UpdateConversationInput
	mockClient := &apitest.MockClient{
		UpdateConversationFunc: func(ctx context.Context, id string, input gen.UpdateConversationInput) (*gen.UpdateConversationResponse, error) {
			capturedInput = input
			return &gen.UpdateConversationResponse{
				UpdateConversation: gen.UpdateConversationUpdateConversation{
					Id: id,
				},
			}, nil
		},
	}

	svc := graphql.NewConversationService(mockClient, logtest.NewScope(t))
	empty := ""
	_, err := svc.Update(context.Background(), "conv-1", graphql.UpdateConversationInput{Title: &empty})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if capturedInput.Title != nil {
		t.Fatalf("Title = %v, want nil", *capturedInput.Title)
	}
	if capturedInput.ClearTitle == nil || !*capturedInput.ClearTitle {
		t.Fatalf("ClearTitle = %v, want true", capturedInput.ClearTitle)
	}
}
