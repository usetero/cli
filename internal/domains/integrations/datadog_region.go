package integrations

// DatadogRegion describes one selectable Datadog site for UI surfaces.
type DatadogRegion struct {
	Site        DatadogSite
	DisplayName string
	Description string
}

var datadogRegions = []DatadogRegion{
	{Site: DatadogSiteUS1, DisplayName: "US1 (datadoghq.com)", Description: "United States"},
	{Site: DatadogSiteUS3, DisplayName: "US3 (us3.datadoghq.com)", Description: "United States"},
	{Site: DatadogSiteUS5, DisplayName: "US5 (us5.datadoghq.com)", Description: "United States"},
	{Site: DatadogSiteUS1Fed, DisplayName: "US1-FED (ddog-gov.com)", Description: "US Government"},
	{Site: DatadogSiteEU1, DisplayName: "EU1 (datadoghq.eu)", Description: "Europe"},
	{Site: DatadogSiteAP1, DisplayName: "AP1 (ap1.datadoghq.com)", Description: "Asia Pacific"},
	{Site: DatadogSiteAP2, DisplayName: "AP2 (ap2.datadoghq.com)", Description: "Asia Pacific"},
}

// DatadogRegions returns the supported Datadog sites in UI display order.
func DatadogRegions() []DatadogRegion {
	out := make([]DatadogRegion, len(datadogRegions))
	copy(out, datadogRegions)
	return out
}
