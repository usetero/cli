package sqlite

import (
	"context"

	"github.com/usetero/cli/internal/sqlite/gen"
)

// CatalogStatus is the aggregated status across all Datadog accounts.
type CatalogStatus struct {
	ServiceCount     int64
	WasteCount       int64
	SavedCount       int64
	AnalyzingCount   int64
	DiscoveringCount int64
	BrokenCount      int64
	EventCount       int64
	WorstStatus      string
	PercentComplete  float64
}

// DatadogAccountStatuses provides access to Datadog account status data.
type DatadogAccountStatuses interface {
	GetCatalogStatus(ctx context.Context) (CatalogStatus, error)
}

// datadogAccountStatusesImpl implements DatadogAccountStatuses.
type datadogAccountStatusesImpl struct {
	queries *gen.Queries
}

// GetCatalogStatus returns aggregated catalog status across all Datadog accounts.
func (d *datadogAccountStatusesImpl) GetCatalogStatus(ctx context.Context) (CatalogStatus, error) {
	row, err := d.queries.GetCatalogStatus(ctx)
	if err != nil {
		return CatalogStatus{}, WrapSQLiteError(err, "get catalog status")
	}

	return CatalogStatus{
		ServiceCount:     row.ServiceCount,
		WasteCount:       row.WasteCount,
		SavedCount:       row.SavedCount,
		AnalyzingCount:   row.AnalyzingCount,
		DiscoveringCount: row.DiscoveringCount,
		BrokenCount:      row.BrokenCount,
		EventCount:       row.EventCount,
		WorstStatus:      row.WorstStatus,
		PercentComplete:  row.PercentComplete,
	}, nil
}
