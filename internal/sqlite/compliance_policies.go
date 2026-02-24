package sqlite

import (
	"context"
	"encoding/json"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/sqlite/gen"
)

// CompliancePolicies provides type-safe access to compliance policies (PII, Secrets, PHI, Payment Data).
type CompliancePolicies interface {
	ListCategorySummaries(ctx context.Context) ([]domain.ComplianceCategorySummary, error)
	ListPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.CompliancePolicy, error)
	CountTotal(ctx context.Context) (int64, error)
	CountFixed(ctx context.Context) (int64, error)
}

// compliancePoliciesImpl implements CompliancePolicies.
type compliancePoliciesImpl struct {
	queries *gen.Queries
}

// ListCategorySummaries returns summary stats for each compliance category.
func (c *compliancePoliciesImpl) ListCategorySummaries(ctx context.Context) ([]domain.ComplianceCategorySummary, error) {
	rows, err := c.queries.ListComplianceCategorySummaries(ctx)
	if err != nil {
		return nil, WrapSQLiteError(err, "list compliance category summaries")
	}

	result := make([]domain.ComplianceCategorySummary, len(rows))
	for i, row := range rows {
		var services []string
		if row.UniqueServices != "" {
			// GROUP_CONCAT returns comma-separated values
			services = splitCommaSeparated(row.UniqueServices)
		}

		category := ""
		if row.Category != nil {
			category = *row.Category
		}

		volumePerHour := 0.0
		if row.VolumePerHour != nil {
			if v, ok := row.VolumePerHour.(float64); ok {
				volumePerHour = v
			}
		}

		result[i] = domain.ComplianceCategorySummary{
			Category:       domain.PolicyCategory(category),
			DisplayName:    row.DisplayName,
			Principle:      row.Principle,
			LeakingCount:   row.LeakingCount,
			AtRiskCount:    row.AtRiskCount,
			FixedCount:     row.FixedCount,
			VolumePerHour:  volumePerHour,
			ServiceCount:   int(row.ServiceCount),
			UniqueServices: services,
		}
	}
	return result, nil
}

// ListPendingPoliciesByCategory returns pending compliance policies for a specific category.
func (c *compliancePoliciesImpl) ListPendingPoliciesByCategory(ctx context.Context, category domain.PolicyCategory, limit int64) ([]domain.CompliancePolicy, error) {
	catStr := string(category)
	rows, err := c.queries.ListPendingCompliancePoliciesByCategory(ctx, gen.ListPendingCompliancePoliciesByCategoryParams{
		Category: &catStr,
		Limit:    limit,
	})
	if err != nil {
		return nil, WrapSQLiteError(err, "list pending compliance policies by category")
	}

	result := make([]domain.CompliancePolicy, 0, len(rows))
	for _, row := range rows {
		p := domain.CompliancePolicy{
			Category:      category,
			LogEventName:  row.LogEventName,
			ServiceName:   row.ServiceName,
			VolumePerHour: row.VolumePerHour,
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

// CountTotal returns the total number of pending compliance policies across all categories.
func (c *compliancePoliciesImpl) CountTotal(ctx context.Context) (int64, error) {
	count, err := c.queries.CountTotalCompliancePolicies(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count total compliance policies")
	}
	return count, nil
}

// CountFixed returns the number of approved compliance policies across all categories.
func (c *compliancePoliciesImpl) CountFixed(ctx context.Context) (int64, error) {
	count, err := c.queries.CountFixedCompliancePolicies(ctx)
	if err != nil {
		return 0, WrapSQLiteError(err, "count fixed compliance policies")
	}
	return count, nil
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

// splitCommaSeparated splits a comma-separated string and deduplicates.
func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var result []string
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && s[j] != ',' {
			j++
		}
		item := s[i:j]
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
		i = j + 1
	}
	return result
}
