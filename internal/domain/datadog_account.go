package domain

// DatadogAccountID is a unique identifier for a Datadog account integration.
type DatadogAccountID string

func (id DatadogAccountID) String() string { return string(id) }

// DatadogSite represents a Datadog datacenter region.
type DatadogSite string

const (
	DatadogSiteUS1    DatadogSite = "US1"
	DatadogSiteUS3    DatadogSite = "US3"
	DatadogSiteUS5    DatadogSite = "US5"
	DatadogSiteEU1    DatadogSite = "EU1"
	DatadogSiteAP1    DatadogSite = "AP1"
	DatadogSiteUS1Fed DatadogSite = "US1_FED"
)

func (s DatadogSite) String() string { return string(s) }

// DatadogRegion contains display information for a Datadog site.
type DatadogRegion struct {
	Site DatadogSite
	Name string
	Desc string
}

// DatadogRegions is the list of available Datadog regions.
var DatadogRegions = []DatadogRegion{
	{DatadogSiteUS1, "US1 (datadoghq.com)", "United States"},
	{DatadogSiteUS3, "US3 (us3.datadoghq.com)", "United States"},
	{DatadogSiteUS5, "US5 (us5.datadoghq.com)", "United States"},
	{DatadogSiteEU1, "EU1 (datadoghq.eu)", "Europe"},
	{DatadogSiteAP1, "AP1 (ap1.datadoghq.com)", "Asia Pacific"},
	{DatadogSiteUS1Fed, "US1-FED (ddog-gov.com)", "US Government"},
}
