//go:build integration

package chat_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/block"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/workos"
	"github.com/usetero/cli/pkg/client"
)

// Integration tests run against real services.
//
// Prerequisites:
//  1. task auth:login
//  2. Control plane running (task dev in control-plane)
//  3. task test:integration

func TestIntegration_MessageService(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger := logtest.New(t)

	// Load config
	cliConfig := config.LoadCLIConfig()

	// Skip if not running against local dev
	if !strings.Contains(cliConfig.ChatEndpoint, "localhost") {
		t.Skip("Skipping: TERO_CHAT_ENDPOINT must point to localhost for integration tests")
	}

	namespace := cliConfig.Namespace()
	if namespace == "" {
		t.Skip("Skipping: not configured for local development")
	}

	t.Logf("Chat Endpoint: %s", cliConfig.ChatEndpoint)

	// Get auth token
	storage := keyring.New(namespace)
	oauthProvider := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.ChatEndpoint, cliConfig.PowerSyncEndpoint)
	authSvc := auth.NewService(oauthProvider, storage, logger)

	token, err := authSvc.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get access token: %v (run: task auth:login)", err)
	}

	// Get account and workspace from preferences
	cfg, err := config.Load(namespace)
	if err != nil {
		t.Fatalf("Config not found: %v (run: task run)", err)
	}
	prefs := preferences.NewService(cfg, logger)

	accountID := prefs.GetDefaultAccountID()
	if accountID == "" {
		t.Fatalf("No default account (run: task run)")
	}

	workspaceID := prefs.GetDefaultWorkspaceID()
	if workspaceID == "" {
		t.Fatalf("No default workspace (run: task run)")
	}

	t.Logf("Account ID: %s", accountID)
	t.Logf("Workspace ID: %s", workspaceID)

	// Create API client for conversation management
	apiClient := client.New(cliConfig.APIEndpoint, token, func() (string, error) {
		return authSvc.GetAccessToken(ctx)
	})
	apiClient.SetAccountID(accountID)
	conversationSvc := api.NewConversationService(apiClient, logger)

	// Create chat client and service
	chatClient := chat.NewClient(cliConfig.ChatEndpoint, logger)
	chatClient.SetToken(token)
	chatClient.SetAccountID(accountID)

	messageSvc := chat.NewMessageService(chatClient)

	t.Run("user message streams response with message_start", func(t *testing.T) {
		// Create a conversation first
		conv, err := conversationSvc.Create(ctx, workspaceID, "Integration Test")
		if err != nil {
			t.Fatalf("Failed to create conversation: %v", err)
		}
		t.Logf("Created conversation: %s", conv.ID)

		// Clean up after test
		defer func() {
			if err := conversationSvc.Delete(ctx, conv.ID); err != nil {
				t.Logf("Warning: failed to delete conversation: %v", err)
			}
		}()

		messageID := uuid.New().String()

		var events []chat.StreamEvent
		var sawMessageStart bool
		var sawDone bool
		var textContent string

		err = messageSvc.UploadUserMessage(ctx, messageID, conv.ID, "Say hello in exactly 3 words", func(event chat.StreamEvent) error {
			events = append(events, event)
			t.Logf("Event: type=%s done=%v raw=%s", event.Type, event.Done, string(event.Raw))

			if event.Type == block.TypeMessageStart {
				sawMessageStart = true
				t.Log("Received message_start")
			}

			if event.Type == block.TypeTextDelta && event.Text != nil {
				textContent += event.Text.Content
			}

			if event.Done {
				sawDone = true
				t.Log("Received done")
			}

			return nil
		})

		if err != nil {
			t.Fatalf("UploadUserMessage error: %v", err)
		}

		if !sawMessageStart {
			t.Error("expected message_start event")
		}

		if !sawDone {
			t.Error("expected done event")
		}

		if textContent == "" {
			t.Error("expected text content in response")
		}

		t.Logf("Received %d events, text: %q", len(events), textContent)
	})
}
