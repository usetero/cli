# Interface Refactor Plan

## Goal
Consolidate all local interfaces to package-defined interfaces. Interfaces live with their implementations, mocks live in `*test` packages.

## Process Per File

For each file with a local interface:

1. **Remove local interface** - Delete the interface definition from the file
2. **Update to package interface** - Change field types and function params to use `api.Accounts`, `auth.Auth`, etc.
3. **Delete local mock** - Remove the mock from the local `*test` package (e.g., `accounttest/`)
4. **Update tests** - Change tests to use mocks from source packages (`apitest.MockAccounts`, `authtest.MockAuth`, etc.)
5. **Rewrite tests if needed** - If tests don't compile or use outdated patterns, rewrite them following TESTING.md
6. **Verify tests pass** - Run `go test` for the package

## Completed

### Mock Packages Created
- [x] `internal/api/apitest/` - MockAccounts, MockOrganizations, MockConversations, MockDatadogAccounts, MockServices, MockClient
- [x] `internal/auth/authtest/` - MockAuth
- [x] `internal/preferences/preferencestest/` - MockPreferences
- [x] `internal/chat/chattest/` - MockMessages

### Package Interfaces Added
- [x] `api.Accounts` - account_service.go
- [x] `api.Organizations` - organization_service.go  
- [x] `api.Conversations` - conversation_service.go
- [x] `api.DatadogAccounts` - datadog_account_service.go
- [x] `api.Services` - service_service.go
- [x] `auth.Auth` - auth_service.go
- [x] `preferences.Preferences` - preferences_service.go
- [x] `powersync.Syncer` - sync.go
- [x] `chat.Messages` - message_service.go

### Files Refactored

#### internal/tui/
- [x] tui.go - uses package interfaces
- [x] tuitest/mock_auth.go, mock_conversations.go, mock_messages.go - DELETED

#### internal/upload/
- [x] upload.go - uses api.Conversations, chat.Messages
- [x] conversation_handler.go - uses api.Conversations
- [x] message_handler.go - uses chat.Messages
- [x] uploadtest/ - DELETED
- [x] upload_test.go - rewritten

#### internal/tui/onboarding/account/
- [x] select.go - uses api.Accounts, preferences.Preferences
- [x] create.go - uses api.Accounts, preferences.Preferences
- [x] default_account_saver.go - DELETED
- [x] accounttest/ - DELETED
- [x] select_test.go - rewritten
- [x] create_test.go - rewritten

#### internal/tui/onboarding/organization/
- [x] select.go - uses api.Organizations, preferences.Preferences, auth.Auth
- [x] create.go - uses api.Organizations, preferences.Preferences, auth.Auth
- [x] default_org_saver.go - DELETED
- [x] organizationtest/ - DELETED
- [x] select_test.go - rewritten
- [x] create_test.go - rewritten

#### internal/tui/onboarding/role/
- [x] select.go - uses preferences.Preferences, auth.Auth (removed RoleSaver, TokenRefresher)

#### internal/tui/onboarding/auth/
- [x] check.go - uses auth.Auth, preferences.Preferences (removed TokenValidator)
- [x] authenticate.go - uses auth.Auth, preferences.Preferences

#### internal/tui/onboarding/
- [x] onboarding.go - uses auth.Auth, preferences.Preferences

#### internal/tui/onboarding/datadog/
- [x] check.go - uses api.DatadogAccounts (removed DatadogAccountChecker)
- [x] api_key.go - uses api.DatadogAccounts (removed DatadogAPIKeyValidator)
- [x] app_key.go - uses api.DatadogAccounts (removed DatadogAccountCreator)
- [x] discovery.go - uses api.DatadogAccounts (removed StatusPoller)
- [x] datadogtest/ - DELETED
- [x] check_test.go - rewritten
- [x] api_key_test.go - rewritten
- [x] app_key_test.go - rewritten
- [x] discovery_test.go - rewritten

## Structural Interfaces (KEEP)

These are UI contracts, not service interfaces. They stay as-is:

- mode/mode.go - Mode interface
- layouts/layout.go - Layout interface
- step/step.go - Step interface
- components/component.go - Component interface
- app/page/page.go - Page interface
- loading/loading.go - SyncState interface

## Summary

All local service interfaces have been removed. The codebase now uses:

- **Package interfaces**: `api.Accounts`, `api.Organizations`, `api.DatadogAccounts`, `auth.Auth`, `preferences.Preferences`
- **Centralized mocks**: `apitest.MockAccounts`, `authtest.MockAuth`, `preferencestest.MockPreferences`

Benefits:
- Single source of truth for each interface
- Consistent mock implementations
- Easier to navigate codebase
- Tests use shared mocks instead of duplicated local mocks
