# CLI

Traditional commands for scripting and automation. Short-lived, queries API directly.

## How It Works

CLI commands are one-shot operations. They don't sync data—they call the GraphQL API directly:

```go
func newDebugStatusCmd() *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // Get authenticated API client
            services, err := getAPIServices(ctx, logger, cliConfig)
            
            // Call API directly
            status, err := services.DatadogAccounts.GetStatus(ctx, accountID)
            
            // Print and exit
            fmt.Printf("Status: %s\n", status.Status)
            return nil
        },
    }
}
```

No PowerSync. No SQLite. Just request → response → exit.

## Current Commands

```
tero                    Launch TUI (default)
tero auth login         Authenticate with device flow
tero auth logout        Clear stored credentials
tero auth status        Show authentication status
tero auth token         Print access token
tero auth switch        Switch organization
tero reset              Clear all local data
tero debug status       Show Datadog account status
tero debug prefs        Show stored preferences
tero debug graphql      Run raw GraphQL query
tero debug paths        Show file paths
```

## Adding Commands

Commands live in `internal/cmd/`. Follow the pattern:

```go
func NewMyCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
    return &cobra.Command{
        Use:   "mycommand",
        Short: "One-line description",
        Long:  "Detailed description.",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Get API services (handles auth)
            services, err := getAPIServices(cmd.Context(), logger, cliConfig)
            if err != nil {
                return err
            }
            
            // 2. Call API
            result, err := services.Something.DoThing(cmd.Context())
            if err != nil {
                return err
            }
            
            // 3. Print output
            fmt.Println(result)
            return nil
        },
    }
}
```

Then add it in `NewRootCmd()`:

```go
func NewRootCmd(logger log.Logger, version string) *cobra.Command {
    rootCmd := newRootCmd(logger, version, cliConfig)
    rootCmd.AddCommand(NewMyCmd(logger, cliConfig))
    return rootCmd
}
```

## API Access Pattern

CLI commands use `getAPIServices()` to get an authenticated client:

```go
func getAPIServices(ctx context.Context, logger log.Logger, cliConfig *config.CLIConfig) (api.APIServices, error) {
    // Set up auth
    tokenStore := keyring.New(namespace)
    workosClient := workos.NewClient(...)
    authService := auth.NewService(workosClient, tokenStore, logger)
    
    // Verify authenticated
    _, err := authService.GetAccessToken(ctx)
    if err != nil {
        return api.APIServices{}, fmt.Errorf("not authenticated: run 'tero auth login' first")
    }
    
    // Return API client
    return api.NewServices(cliConfig.APIEndpoint+"/graphql", authService, logger), nil
}
```

This gives you type-safe access to the GraphQL API with automatic token refresh.

## Output Styling

Use the theme for consistent output:

```go
theme := styles.NewTheme(true)
s := theme.Styles

fmt.Println(s.Title.Render("Section Title"))
fmt.Println(s.Help.Render("Label:"), s.Body.Render(value))
fmt.Println(s.Success.Render("✓ Done"))
fmt.Println(s.Error.Render("✗ Failed"))
```

## Future Commands

As we expand, we'll add commands like:

```
tero services list      List services from catalog
tero services checkout  Check out a service
tero policies list      List policies
```

These will follow the same pattern: call API directly, print, exit.

## Code Location

```
internal/cmd/
├── root.go         Root command, wiring
├── execute.go      Entry point
├── auth.go         Auth subcommands
├── debug.go        Debug subcommands
└── reset.go        Reset command
```
