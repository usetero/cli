package syncer

type lifecyclePhase string

const (
	phaseDisconnected lifecyclePhase = "disconnected"
	phaseConnecting   lifecyclePhase = "connecting"
	phaseSyncing      lifecyclePhase = "syncing"
	phaseReady        lifecyclePhase = "ready"
	phaseReconnecting lifecyclePhase = "reconnecting"
	phaseError        lifecyclePhase = "error"
)

func (s *Syncer) transitionLocked(phase lifecyclePhase, state State) {
	s.phase = phase
	s.state = state
}

func (s *Syncer) transition(phase lifecyclePhase, state State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transitionLocked(phase, state)
}

func (s *Syncer) transitionToConnectingLocked() {
	s.transitionLocked(phaseConnecting, &Connecting{})
}

func (s *Syncer) transitionToDisconnectedLocked() {
	s.transitionLocked(phaseDisconnected, &Disconnected{})
}

func (s *Syncer) transitionToSyncing(progress *Progress) {
	s.transition(phaseSyncing, &Syncing{Progress: progress})
}

func (s *Syncer) transitionToReady() {
	s.transition(phaseReady, &Ready{})
}

func (s *Syncer) transitionToReconnecting(retries int) {
	s.transition(phaseReconnecting, &Reconnecting{Degraded: retries >= s.retry.ErrorStateAfter})
}

func (s *Syncer) transitionToError(err error) {
	s.transition(phaseError, &Error{Err: err})
}
