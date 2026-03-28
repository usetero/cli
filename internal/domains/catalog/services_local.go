package catalog

import (
	"context"
	"database/sql"

	servicesdb "github.com/usetero/cli/internal/domains/catalog/db/servicesgen"
)

// LocalCatalogService uses SQLite/sqlc for local catalog reads.
type LocalCatalogService struct {
	services  *servicesdb.Queries
	logEvents LogEventService
}

func NewLocalCatalogService(db *sql.DB) *LocalCatalogService {
	if db == nil {
		panic("catalog local service requires db")
	}
	return &LocalCatalogService{
		services:  servicesdb.New(db),
		logEvents: NewLocalLogEventService(db),
	}
}

func (s *LocalCatalogService) ListServices(ctx context.Context) ([]Service, error) {
	rows, err := s.services.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	factRows, err := s.services.ListAllServiceFacts(ctx)
	if err != nil {
		return nil, err
	}

	factsByService, err := decodeServiceFactsByService(factRows)
	if err != nil {
		return nil, err
	}

	out := make([]Service, 0, len(rows))
	for _, row := range rows {
		service := fromServicesDBListRow(row)
		service.Facts = factsByService[service.ID]
		out = append(out, service)
	}
	return out, nil
}

func (s *LocalCatalogService) ListLogEventsByService(ctx context.Context, serviceID ServiceID) ([]LogEvent, error) {
	return s.logEvents.ListLogEventsByService(ctx, serviceID)
}

func decodeServiceFactsByService(rows []servicesdb.ListAllServiceFactsRow) (map[ServiceID]ServiceFacts, error) {
	rawByService := make(map[ServiceID][]ServiceFactsRow)
	for _, row := range rows {
		serviceID := ServiceID(value(row.ServiceID))
		rawByService[serviceID] = append(rawByService[serviceID], fromServicesDBFactRow(row))
	}

	out := make(map[ServiceID]ServiceFacts, len(rawByService))
	for serviceID, rawRows := range rawByService {
		facts, err := DecodeServiceFacts(rawRows)
		if err != nil {
			return nil, err
		}
		out[serviceID] = facts
	}
	return out, nil
}
