package accounts

import "github.com/usetero/cli/internal/domain"

func findAccountByID(accounts []domain.Account, id domain.AccountID) *domain.Account {
	if id == "" {
		return nil
	}
	for _, account := range accounts {
		if account.ID == id {
			resolved := account
			return &resolved
		}
	}
	return nil
}
