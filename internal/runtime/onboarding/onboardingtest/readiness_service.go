package onboardingtest

import "context"

type ReadinessService struct {
	ReadyValue bool
}

func (s ReadinessService) Ready(context.Context) (bool, error) { return s.ReadyValue, nil }
