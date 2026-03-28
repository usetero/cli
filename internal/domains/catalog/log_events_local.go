package catalog

import (
	"context"
	"database/sql"
	"fmt"

	logeventsdb "github.com/usetero/cli/internal/domains/catalog/db/logeventsgen"
)

// LocalLogEventService uses SQLite/sqlc for local log-event reads.
type LocalLogEventService struct {
	q *logeventsdb.Queries
}

func NewLocalLogEventService(db *sql.DB) *LocalLogEventService {
	if db == nil {
		panic("catalog local log event service requires db")
	}
	return &LocalLogEventService{q: logeventsdb.New(db)}
}

func (s *LocalLogEventService) ListLogEventsByService(ctx context.Context, serviceID ServiceID) ([]LogEvent, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service id is required")
	}
	rows, err := s.q.ListLogEventsByService(ctx, ptrString(string(serviceID)))
	if err != nil {
		return nil, err
	}
	factRows, err := s.q.ListLogEventFactsByService(ctx, ptrString(string(serviceID)))
	if err != nil {
		return nil, err
	}

	factsByEvent, err := decodeLogEventFactsByID(factRows)
	if err != nil {
		return nil, err
	}

	out := make([]LogEvent, 0, len(rows))
	for _, row := range rows {
		event := fromLogEventsDBListByServiceRow(row)
		event.Facts = factsByEvent[event.ID]
		out = append(out, event)
	}
	return out, nil
}

func decodeLogEventFactsByID(rows []logeventsdb.ListLogEventFactsByServiceRow) (map[LogEventID]LogEventFacts, error) {
	rawByEvent := make(map[LogEventID][]LogEventFactsRow)
	for _, row := range rows {
		logEventID := LogEventID(value(row.LogEventID))
		rawByEvent[logEventID] = append(rawByEvent[logEventID], fromLogEventsDBFactByServiceRow(row))
	}

	out := make(map[LogEventID]LogEventFacts, len(rawByEvent))
	for logEventID, rawRows := range rawByEvent {
		facts, err := DecodeLogEventFacts(rawRows)
		if err != nil {
			return nil, err
		}
		out[logEventID] = facts
	}
	return out, nil
}
