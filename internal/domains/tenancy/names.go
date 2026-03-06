package tenancy

import (
	"strings"

	"github.com/usetero/cli/internal/domains/validation"
)

type OrganizationName string

func (n OrganizationName) String() string {
	return string(n)
}

func ParseOrganizationName(raw string) (OrganizationName, error) {
	name := OrganizationName(strings.TrimSpace(raw))
	if err := validation.Struct(struct {
		Name string `label:"organization name" validate:"required,notblank,max=100"`
	}{Name: name.String()}); err != nil {
		return "", err
	}
	return name, nil
}
