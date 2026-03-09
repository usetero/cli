package terotest

import "time"

type Options struct {
	Args         []string
	Env          map[string]string
	HomeDir      string
	StopTimeout  time.Duration
	WindowWidth  uint16
	WindowHeight uint16
}
