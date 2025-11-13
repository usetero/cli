package datadog

// GetAPIKeyURL returns the region-specific URL for creating API keys
func GetAPIKeyURL(site string) string {
	domain := getDomainForSite(site)
<<<<<<< HEAD
	return "https://app." + domain + "/organization-settings/api-keys"
=======
	// US1 uses app. subdomain, all others use the domain directly
	if site == "US1" {
		return "https://app." + domain + "/organization-settings/api-keys"
	}
	return "https://" + domain + "/organization-settings/api-keys"
>>>>>>> 17e8dd9 (chore: initial commit)
}

// GetAppKeyURL returns the region-specific URL for creating service accounts
func GetAppKeyURL(site string) string {
	domain := getDomainForSite(site)
<<<<<<< HEAD
	return "https://app." + domain + "/organization-settings/service-accounts"
=======
	// US1 uses app. subdomain, all others use the domain directly
	if site == "US1" {
		return "https://app." + domain + "/organization-settings/service-accounts"
	}
	return "https://" + domain + "/organization-settings/service-accounts"
>>>>>>> 17e8dd9 (chore: initial commit)
}

// getDomainForSite returns the domain for a given Datadog site code
func getDomainForSite(site string) string {
	for _, r := range regions {
		if r.site == site {
			return r.domain
		}
	}
	// Default to US1 if unknown
	return "datadoghq.com"
}
