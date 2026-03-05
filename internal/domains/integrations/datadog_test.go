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
