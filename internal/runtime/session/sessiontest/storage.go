package sessiontest

import (
	"github.com/usetero/cli/internal/infrastructure/sqlite"
)

type Storage struct {
	Path sqlite.DatabasePath
	Err  error
}

func (s Storage) DatabasePath(sqlite.AccountID) (sqlite.DatabasePath, error) {
	return s.Path, s.Err
}
