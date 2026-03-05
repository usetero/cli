package uploadertest

import (
	"context"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
)

type TokenSource struct {
	Token psclient.AccessToken
}

func (s *TokenSource) GetAccessToken(context.Context) (psclient.AccessToken, error) {
	return s.Token, nil
}
