package change

import "time"

type Status struct {
	IsolatedEventsPerHour   *float64
	IsolatedTotalUSDPerHour *float64
	LogEventCount           *int64
	FindingUpdatedAt        time.Time
}
