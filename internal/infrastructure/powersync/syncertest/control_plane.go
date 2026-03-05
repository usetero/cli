package syncertest

import (
	"context"
	"sync/atomic"

	"github.com/usetero/cli/internal/infrastructure/powersync/extension"
)

type ControlPlane struct {
	StartInstructions        []extension.Instruction
	SendTextInstructions     []extension.Instruction
	NotifyUploadInstructions []extension.Instruction
	SendTextLineHook         func()
	NotifyUploadHook         func()

	NotifyTokenRefreshedCalls atomic.Int32
	ActiveCalls               atomic.Int32
	MaxConcurrentCalls        atomic.Int32
}

func (c *ControlPlane) Start(context.Context, extension.StartRequest) ([]extension.Instruction, error) {
	return c.StartInstructions, nil
}

func (c *ControlPlane) SendTextLine(context.Context, string) ([]extension.Instruction, error) {
	c.enter()
	defer c.leave()
	if c.SendTextLineHook != nil {
		c.SendTextLineHook()
	}
	return c.SendTextInstructions, nil
}

func (c *ControlPlane) NotifyConnection(context.Context, extension.ConnectionEvent) ([]extension.Instruction, error) {
	return nil, nil
}

func (c *ControlPlane) NotifyTokenRefreshed(context.Context) ([]extension.Instruction, error) {
	c.NotifyTokenRefreshedCalls.Add(1)
	return nil, nil
}

func (c *ControlPlane) NotifyUploadCompleted(context.Context) ([]extension.Instruction, error) {
	c.enter()
	defer c.leave()
	if c.NotifyUploadHook != nil {
		c.NotifyUploadHook()
	}
	return c.NotifyUploadInstructions, nil
}

func (c *ControlPlane) Close() error { return nil }

func (c *ControlPlane) enter() {
	active := c.ActiveCalls.Add(1)
	for {
		max := c.MaxConcurrentCalls.Load()
		if active <= max {
			return
		}
		if c.MaxConcurrentCalls.CompareAndSwap(max, active) {
			return
		}
	}
}

func (c *ControlPlane) leave() {
	c.ActiveCalls.Add(-1)
}
