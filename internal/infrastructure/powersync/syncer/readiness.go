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
	if runtime == nil {
		panic("powersync runtime readiness requires runtime")
	}
	return RuntimeReadiness{runtime: runtime}
}

func (r RuntimeReadiness) Ready(_ context.Context) (bool, error) {
	return r.runtime.IsReady(), nil
}
