package sqlite

import (
	"context"
	"encoding/json"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// CompliancePolicies provides type-safe access to compliance policies (PII, Secrets, PHI, Payment Data).
type CompliancePolicies interface {
	ListPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.CompliancePolicy, error)
}

// compliancePoliciesImpl implements CompliancePolicies.
type compliancePoliciesImpl struct {
	queries *gen.Queries
}

// ListPendingPoliciesByCategory returns pending compliance policies for a specific category.
func (c *compliancePoliciesImpl) ListPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.CompliancePolicy, error) {
	_ = limit
	rows, err := c.queries.ListPendingCompliancePoliciesByCategory(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list pending compliance policies by category")
	}

	result := make([]domain.CompliancePolicy, 0, len(rows))
	for _, row := range rows {
		p := domain.CompliancePolicy{
			Category:      category,
			LogEventName:  row.LogEventName,
			ServiceName:   row.ServiceName,
			VolumePerHour: float64Ptr(row.VolumePerHour),
			AnyObserved:   row.AnyObserved != 0,
		}

		// Parse the analysis JSON to extract sensitive fields for this category.
		if row.Analysis != "" {
			fields := extractSensitiveFields(row.Analysis, category)
			p.Fields = fields
		}

		result = append(result, p)
	}
	return result, nil
}

// extractSensitiveFields parses the analysis JSON and extracts the fields array for the given category.
func extractSensitiveFields(analysisJSON string, category domain.PolicyCategory) []domain.SensitiveField {
	var envelope domain.ComplianceAnalysisEnvelope
	if err := json.Unmarshal([]byte(analysisJSON), &envelope); err != nil {
		return nil
	}

	switch category {
	case domain.CategoryPIILeakage:
		if envelope.PIILeakage != nil {
			return envelope.PIILeakage.Fields
		}
	case domain.CategorySecretsLeakage:
		if envelope.SecretsLeakage != nil {
			return envelope.SecretsLeakage.Fields
		}
	case domain.CategoryPHILeakage:
		if envelope.PHILeakage != nil {
			return envelope.PHILeakage.Fields
		}
	case domain.CategoryPaymentDataLeakage:
		if envelope.PaymentDataLeakage != nil {
			return envelope.PaymentDataLeakage.Fields
		}
	default:
		return nil
	}

	return nil
}
