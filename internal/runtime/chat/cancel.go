package chat

func (r *Runtime) Cancel() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel == nil {
		return nil
	}
	r.cancel()
	r.cancel = nil
	return nil
}
