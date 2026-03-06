package screen

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// LiftFunc can transform child commands before bubbling to parent.
// Typical use: map child message types into parent message types.
type LiftFunc func(cmd tea.Cmd) tea.Cmd

type childSlot struct {
	model  Model
	active bool
	lift   LiftFunc
}

// Router manages child models and forwards messages to active children.
//
// It supports both:
// - mutually exclusive routing via ActivateOnly
// - multi-active routing via SetActive on multiple children.
type Router[ID ~string] struct {
	order []ID
	slots map[ID]*childSlot
}

func (r *Router[ID]) ensureSlots() {
	if r.slots == nil {
		r.slots = map[ID]*childSlot{}
	}
}

// Register adds or replaces a named child model.
func (r *Router[ID]) Register(name ID, model Model) {
	if name == "" {
		panic("child name is required")
	}
	if model == nil {
		panic("child model is required")
	}
	r.ensureSlots()
	if slot, ok := r.slots[name]; ok {
		slot.model = model
		return
	}
	r.order = append(r.order, name)
	r.slots[name] = &childSlot{model: model}
}

// SetModel replaces a registered child model. Returns false when child does not exist.
func (r *Router[ID]) SetModel(name ID, model Model) bool {
	if model == nil {
		panic("child model is required")
	}
	slot, ok := r.slots[name]
	if !ok {
		return false
	}
	slot.model = model
	return true
}

// Model returns a registered child model, if present.
func (r *Router[ID]) Model(name ID) Model {
	slot, ok := r.slots[name]
	if !ok {
		return nil
	}
	return slot.model
}

// SetLift sets optional command lifting for a child. Returns false when child does not exist.
func (r *Router[ID]) SetLift(name ID, lift LiftFunc) bool {
	slot, ok := r.slots[name]
	if !ok {
		return false
	}
	slot.lift = lift
	return true
}

// SetActive toggles active state for a child. Returns false when child does not exist.
func (r *Router[ID]) SetActive(name ID, active bool) bool {
	slot, ok := r.slots[name]
	if !ok {
		return false
	}
	slot.active = active
	return true
}

// ActivateOnly enables one child and disables all others. Returns false when child does not exist.
func (r *Router[ID]) ActivateOnly(name ID) bool {
	if _, ok := r.slots[name]; !ok {
		return false
	}
	for _, childName := range r.order {
		slot := r.slots[childName]
		slot.active = childName == name
	}
	return true
}

// ClearActive disables all children.
func (r *Router[ID]) ClearActive() {
	for _, name := range r.order {
		r.slots[name].active = false
	}
}

// Forward sends msg to every active child in registration order.
func (r *Router[ID]) Forward(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(r.order))
	for _, name := range r.order {
		slot := r.slots[name]
		if slot == nil || !slot.active || slot.model == nil {
			continue
		}
		next, cmd := slot.model.Update(msg)
		if model, ok := next.(Model); ok {
			slot.model = model
		}
		if slot.lift != nil {
			cmd = slot.lift(cmd)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// SetSize forwards size to active children.
func (r *Router[ID]) SetSize(width, height int) {
	for _, name := range r.order {
		slot := r.slots[name]
		if slot == nil || !slot.active || slot.model == nil {
			continue
		}
		slot.model.SetSize(width, height)
	}
}

// SetSizeAll forwards size to all registered children.
func (r *Router[ID]) SetSizeAll(width, height int) {
	for _, name := range r.order {
		slot := r.slots[name]
		if slot == nil || slot.model == nil {
			continue
		}
		slot.model.SetSize(width, height)
	}
}

// ShortHelp returns merged key bindings from all active children.
func (r *Router[ID]) ShortHelp() []key.Binding {
	var bindings []key.Binding
	for _, name := range r.order {
		slot := r.slots[name]
		if slot == nil || !slot.active || slot.model == nil {
			continue
		}
		bindings = append(bindings, slot.model.ShortHelp()...)
	}
	return bindings
}
