package cmd

import (
	"context"
	"fmt"

	"github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/workos"
)

func newAuthService(cliConfig *config.CLIConfig, scope log.Scope) *auth.Service {
	env := cliConfig.Environment()
	tokenStore := keyring.New(env)
	workosClient := workos.NewClient(
		cliConfig.WorkOSClientID,
		cliConfig.APIEndpoint,
		cliConfig.PowerSyncEndpoint,
		cliConfig.ChatEndpoint,
	)
	return auth.NewService(workosClient, tokenStore, scope)
}

func newGraphQLServiceSet(cliConfig *config.CLIConfig, authService auth.Auth, scope log.Scope) graphql.ServiceSet {
	return graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authService, scope)
}

func newAuthenticatedGraphQLServiceSet(ctx context.Context, cliConfig *config.CLIConfig, scope log.Scope) (graphql.ServiceSet, error) {
	authService := newAuthService(cliConfig, scope)
	if _, err := authService.GetAccessToken(ctx); err != nil {
		return graphql.ServiceSet{}, fmt.Errorf("not authenticated: run 'tero auth login' first")
	}
	return newGraphQLServiceSet(cliConfig, authService, scope), nil
}
