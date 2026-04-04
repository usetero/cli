package terotest

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/usetero/cli/internal/infrastructure/auth/keyring"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

const (
	envE2EAccessToken   = "TERO_E2E_ACCESS_TOKEN"
	envE2ERefreshToken  = "TERO_E2E_REFRESH_TOKEN"
	envE2EDatadogAPIKey = "TERO_E2E_DATADOG_API_KEY"
	envE2EDatadogAppKey = "TERO_E2E_DATADOG_APP_KEY"
)

type LocalRun struct {
	RunID            string
	OrganizationName string
	DatadogName      string
	DatadogAPIKey    string
	DatadogAppKey    string
	AppEnv           map[string]string
	apiOrigin        string
	accessToken      string
}

func RequireLocalRun(t testing.TB, homeDir string) LocalRun {
	t.Helper()

	cfg, err := config.Resolve(config.RuntimeConfig{Env: config.ProfileLocal})
	if err != nil {
		t.Fatalf("resolve local runtime config: %v", err)
	}

	apiOrigin := getenvDefault("TERO_API_ORIGIN", cfg.API.Origin)
	chatOrigin := getenvDefault("TERO_CHAT_ORIGIN", cfg.Chat.Origin)
	powerSyncOrigin := getenvDefault("TERO_POWERSYNC_ORIGIN", cfg.PowerSync.Origin)

	requireReachable(t, "control plane", apiOrigin)
	requireReachable(t, "chat", chatOrigin)
	requireReachable(t, "powersync", powerSyncOrigin)

	accessToken, refreshToken := resolveTokens(t)
	SeedTokens(t, homeDir, "local", accessToken, refreshToken)

	runID := strings.ToLower(fmt.Sprintf("%d", time.Now().UnixNano()))
	return LocalRun{
		RunID:            runID,
		OrganizationName: "tero-e2e-" + runID,
		DatadogName:      "tero-e2e-" + runID,
		DatadogAPIKey:    requiredEnv(t, envE2EDatadogAPIKey),
		DatadogAppKey:    requiredEnv(t, envE2EDatadogAppKey),
		apiOrigin:        apiOrigin,
		accessToken:      accessToken,
		AppEnv: map[string]string{
			"TERO_ENV":              "local",
			keyring.EnvBackend:      keyring.BackendFile,
			keyring.EnvPath:         SecretStorePath(homeDir, "local"),
			"TERO_API_ORIGIN":       apiOrigin,
			"TERO_CHAT_ORIGIN":      chatOrigin,
			"TERO_POWERSYNC_ORIGIN": powerSyncOrigin,
		},
	}
}

func (r LocalRun) APIClient() *controlplane.BootstrapClient {
	return controlplane.NewBootstrapClient(r.apiOrigin, staticTokenProvider(r.accessToken))
}

func (r LocalRun) FindOrganizationID(ctx context.Context, name string) (controlplane.OrganizationID, error) {
	orgs, err := r.APIClient().ListOrganizations(ctx)
	if err != nil {
		return "", err
	}
	for _, org := range orgs {
		if org.Name == name {
			return org.ID, nil
		}
	}
	return "", fmt.Errorf("organization %q not found", name)
}

func (r LocalRun) DeleteOrganization(ctx context.Context, id controlplane.OrganizationID) error {
	return r.APIClient().DeleteOrganization(ctx, id)
}

func resolveTokens(t testing.TB) (string, string) {
	t.Helper()

	accessToken := strings.TrimSpace(os.Getenv(envE2EAccessToken))
	refreshToken := strings.TrimSpace(os.Getenv(envE2ERefreshToken))
	switch {
	case accessToken != "" && refreshToken != "":
		return accessToken, refreshToken
	case accessToken != "" || refreshToken != "":
		t.Fatalf("set both %s and %s or neither", envE2EAccessToken, envE2ERefreshToken)
	}

	store, err := openSystemStore("local")
	if err != nil {
		t.Fatalf("open local auth store: %v", err)
	}
	accessToken, err = store.Get(keyring.KeyAccessToken)
	if err != nil {
		t.Fatalf("load local access token: %v", err)
	}
	refreshToken, err = store.Get(keyring.KeyRefreshToken)
	if err != nil {
		t.Fatalf("load local refresh token: %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("local auth tokens are required; authenticate first or set %s and %s", envE2EAccessToken, envE2ERefreshToken)
	}
	return accessToken, refreshToken
}

func openSystemStore(env string) (*keyring.Store, error) {
	prevBackend, hadBackend := os.LookupEnv(keyring.EnvBackend)
	prevPath, hadPath := os.LookupEnv(keyring.EnvPath)
	_ = os.Unsetenv(keyring.EnvBackend)
	_ = os.Unsetenv(keyring.EnvPath)
	defer func() {
		if hadBackend {
			_ = os.Setenv(keyring.EnvBackend, prevBackend)
		} else {
			_ = os.Unsetenv(keyring.EnvBackend)
		}
		if hadPath {
			_ = os.Setenv(keyring.EnvPath, prevPath)
		} else {
			_ = os.Unsetenv(keyring.EnvPath)
		}
	}()
	return keyring.NewStore(env)
}

func requireReachable(t testing.TB, name, origin string) {
	t.Helper()

	u, err := url.Parse(origin)
	if err != nil {
		t.Fatalf("parse %s origin %q: %v", name, origin, err)
	}
	address := u.Host
	if !strings.Contains(address, ":") {
		switch u.Scheme {
		case "https":
			address += ":443"
		default:
			address += ":80"
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		t.Fatalf("%s at %s is unreachable: %v", name, origin, err)
	}
	_ = conn.Close()
}

func requiredEnv(t testing.TB, key string) string {
	t.Helper()

	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		t.Fatalf("%s is required for local onboarding E2E", key)
	}
	return value
}

func getenvDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

type staticTokenProvider string

func (p staticTokenProvider) GetAccessToken(context.Context) (string, error) {
	return string(p), nil
}
