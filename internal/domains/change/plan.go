package change

import "time"

type Plan struct {
	Status        string
	Title         string
	Summary       string
	Rationale     string
	OpenQuestions string
	Steps         string
	Revision      int64
	Version       int64
	UpdatedAt     time.Time
}
