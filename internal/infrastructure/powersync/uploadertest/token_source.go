package uploadertest

import (
	"context"

	psclient "github.com/usetero/cli/internal/infrastructure/powersync/client"
	psuploader "github.com/usetero/cli/internal/infrastructure/powersync/uploader"
)

type TokenSource struct {
	Token psclient.AccessToken
}

var _ psuploader.TokenSource = (*TokenSource)(nil)

func (s *TokenSource) GetAccessToken(context.Context) (psclient.AccessToken, error) {
	return s.Token, nil
}
