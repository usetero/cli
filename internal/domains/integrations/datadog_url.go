package integrations

// DatadogAPIKeyURL returns the region-specific page for creating API keys.
func DatadogAPIKeyURL(site DatadogSite) string {
	domain := datadogDomain(site)
	if site == DatadogSiteUS1 {
		return "https://app." + domain + "/organization-settings/api-keys"
	}
	return "https://" + domain + "/organization-settings/api-keys"
}

// DatadogAppKeyURL returns the region-specific page for creating service
// account application keys.
func DatadogAppKeyURL(site DatadogSite) string {
	domain := datadogDomain(site)
	if site == DatadogSiteUS1 {
		return "https://app." + domain + "/organization-settings/service-accounts"
	}
	return "https://" + domain + "/organization-settings/service-accounts"
}

func datadogDomain(site DatadogSite) string {
	switch site {
	case DatadogSiteUS1:
		return "datadoghq.com"
	case DatadogSiteUS3:
		return "us3.datadoghq.com"
	case DatadogSiteUS5:
		return "us5.datadoghq.com"
	case DatadogSiteEU1:
		return "datadoghq.eu"
	case DatadogSiteUS1Fed:
		return "ddog-gov.com"
	case DatadogSiteAP1:
		return "ap1.datadoghq.com"
	case DatadogSiteAP2:
		return "ap2.datadoghq.com"
	default:
		return "datadoghq.com"
	}
}
