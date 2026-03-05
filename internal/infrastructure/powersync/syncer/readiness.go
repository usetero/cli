package syncer

import "context"

// ReadinessService reports whether the current sync runtime is ready.
type ReadinessService interface {
	Ready(ctx context.Context) (bool, error)
}

// RuntimeReadiness adapts a syncer runtime to the ReadinessService contract.
type RuntimeReadiness struct {
	runtime interface{ IsReady() bool }
}

func NewRuntimeReadiness(runtime interface{ IsReady() bool }) RuntimeReadiness {
	return RuntimeReadiness{runtime: runtime}
}

func (r RuntimeReadiness) Ready(_ context.Context) (bool, error) {
	if r.runtime == nil {
		return false, nil
	}
	return r.runtime.IsReady(), nil
}
