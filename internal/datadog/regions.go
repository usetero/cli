package datadog

// region represents a Datadog site/region with its API domain and enum identifier
type region struct {
	domain string // API domain (e.g., "datadoghq.com")
	site   string // GraphQL enum value (e.g., "US1")
}

// Common Datadog regions in order of popularity.
var regions = []region{
	{domain: "datadoghq.com", site: "US1"},     // US1 (most common)
	{domain: "us5.datadoghq.com", site: "US5"}, // US5
	{domain: "us3.datadoghq.com", site: "US3"}, // US3
	{domain: "datadoghq.eu", site: "EU1"},      // EU1
	{domain: "ap1.datadoghq.com", site: "AP1"}, // AP1
	{domain: "ap2.datadoghq.com", site: "AP2"}, // AP2
	{domain: "ddog-gov.com", site: "US1_FED"},  // US1 FedRAMP
}

// Region represents a Datadog region for UI display
type Region struct {
	Site        string // GraphQL enum value (US1, US5, EU1, etc.)
	Domain      string // API domain (e.g., "datadoghq.com")
	DisplayName string // Human-readable name
}

// GetRegions returns the list of available Datadog regions for UI display
func GetRegions() []Region {
	result := make([]Region, len(regions))
	for i, r := range regions {
		displayName := r.site
		if r.site == "US1_FED" {
			displayName = "US1 FedRAMP"
		}
		result[i] = Region{
			Site:        r.site,
			Domain:      r.domain,
			DisplayName: displayName,
		}
	}
	return result
}
