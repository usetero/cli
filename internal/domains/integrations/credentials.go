package integrations

import (
	"strings"

	"github.com/usetero/cli/internal/domains/validation"
)

type DatadogAccountName string

func ParseDatadogAccountName(raw string) (DatadogAccountName, error) {
	name := DatadogAccountName(strings.TrimSpace(raw))
	if err := validation.Struct(struct {
		Name string `label:"datadog account name" validate:"required,notblank,max=100"`
	}{Name: name.String()}); err != nil {
		return "", err
	}
	return name, nil
}

func (n DatadogAccountName) String() string { return string(n) }

type DatadogAPIKey string

func ParseDatadogAPIKey(raw string) (DatadogAPIKey, error) {
	key := DatadogAPIKey(strings.TrimSpace(raw))
	if err := validation.Struct(struct {
		Key string `label:"datadog api key" validate:"required,notblank"`
	}{Key: key.String()}); err != nil {
		return "", err
	}
	return key, nil
}

func (k DatadogAPIKey) String() string { return string(k) }

type DatadogAppKey string

func ParseDatadogAppKey(raw string) (DatadogAppKey, error) {
	key := DatadogAppKey(strings.TrimSpace(raw))
	if err := validation.Struct(struct {
		Key string `label:"datadog app key" validate:"required,notblank"`
	}{Key: key.String()}); err != nil {
		return "", err
	}
	return key, nil
}

func (k DatadogAppKey) String() string { return string(k) }
