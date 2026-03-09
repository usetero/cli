package integrations

import "testing"

func TestDatadogSite_Valid(t *testing.T) {
	validSites := []DatadogSite{
		DatadogSiteUS1,
		DatadogSiteUS3,
		DatadogSiteUS5,
		DatadogSiteEU1,
		DatadogSiteUS1Fed,
		DatadogSiteAP1,
		DatadogSiteAP2,
	}
	for _, site := range validSites {
		if !site.Valid() {
			t.Fatalf("expected site %q to be valid", site)
		}
	}

	if DatadogSite("INVALID").Valid() {
		t.Fatalf("expected invalid site to be rejected")
	}
}

func TestDatadogRegions(t *testing.T) {
	regions := DatadogRegions()
	if len(regions) != 7 {
		t.Fatalf("expected 7 regions, got %d", len(regions))
	}

	expected := []DatadogRegion{
		{Site: DatadogSiteUS1, DisplayName: "US1 (datadoghq.com)", Description: "United States"},
		{Site: DatadogSiteUS3, DisplayName: "US3 (us3.datadoghq.com)", Description: "United States"},
		{Site: DatadogSiteUS5, DisplayName: "US5 (us5.datadoghq.com)", Description: "United States"},
		{Site: DatadogSiteEU1, DisplayName: "EU1 (datadoghq.eu)", Description: "Europe"},
		{Site: DatadogSiteAP1, DisplayName: "AP1 (ap1.datadoghq.com)", Description: "Asia Pacific"},
		{Site: DatadogSiteUS1Fed, DisplayName: "US1-FED (ddog-gov.com)", Description: "US Government"},
		{Site: DatadogSiteAP2, DisplayName: "AP2 (ap2.datadoghq.com)", Description: "Asia Pacific"},
	}
	for i := range expected {
		if regions[i] != expected[i] {
			t.Fatalf("unexpected region at %d: got %+v want %+v", i, regions[i], expected[i])
		}
	}

	regions[0].DisplayName = "mutated"
	fresh := DatadogRegions()
	if fresh[0].DisplayName != expected[0].DisplayName {
		t.Fatalf("expected returned regions to be copied, got %q", fresh[0].DisplayName)
	}
}

func TestDatadogURLs(t *testing.T) {
	if got := DatadogAPIKeyURL(DatadogSiteUS1); got != "https://app.datadoghq.com/organization-settings/api-keys" {
		t.Fatalf("unexpected US1 api key url: %q", got)
	}
	if got := DatadogAPIKeyURL(DatadogSiteEU1); got != "https://datadoghq.eu/organization-settings/api-keys" {
		t.Fatalf("unexpected EU1 api key url: %q", got)
	}
	if got := DatadogAppKeyURL(DatadogSiteUS1); got != "https://app.datadoghq.com/organization-settings/service-accounts" {
		t.Fatalf("unexpected US1 app key url: %q", got)
	}
	if got := DatadogAppKeyURL(DatadogSiteUS1Fed); got != "https://ddog-gov.com/organization-settings/service-accounts" {
		t.Fatalf("unexpected US1_FED app key url: %q", got)
	}
}
