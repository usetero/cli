package chat

func cloneState(in State) State {
	out := in
	out.Messages = append([]MessageView(nil), in.Messages...)
	return out
}

func (r *Runtime) publish(state State) {
	select {
	case r.updates <- state:
	default:
		select {
		case <-r.updates:
		default:
		}
		select {
		case r.updates <- state:
		default:
		}
	}
}

func (r *Runtime) publishLocked() {
	r.publish(cloneState(r.state))
}
