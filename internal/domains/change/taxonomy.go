package change

import "strings"

type DomainDescriptor struct {
	Key         Domain
	Label       string
	Description string
}

type TypeDescriptor struct {
	Domain      Domain
	Dimension   string
	Key         string
	Label       string
	Description string
}

var domainDescriptors = map[Domain]DomainDescriptor{
	"quality": {
		Key:         "quality",
		Label:       "Quality",
		Description: "Telemetry quality findings that improve signal fidelity, schema clarity, and operator trust.",
	},
	"compliance": {
		Key:         "compliance",
		Label:       "Compliance",
		Description: "Sensitive-data and governance findings that reduce exposure and operational risk.",
	},
}

var typeDescriptors = map[string]TypeDescriptor{
	"wrong_severity": {
		Domain:      "quality",
		Dimension:   "signal_fidelity",
		Key:         "wrong_severity",
		Label:       "Wrong severity",
		Description: "Severity no longer matches the impact or urgency of the event.",
	},
	"schema_drift": {
		Domain:      "quality",
		Dimension:   "schema_clarity",
		Key:         "schema_drift",
		Label:       "Schema drift",
		Description: "The event structure is drifting in ways that make downstream use inconsistent.",
	},
	"traceability_gap": {
		Domain:      "quality",
		Dimension:   "traceability",
		Key:         "traceability_gap",
		Label:       "Traceability gap",
		Description: "Important identifiers or linking context are missing from operational records.",
	},
	"sensitive_data_exposure": {
		Domain:      "compliance",
		Dimension:   "sensitive_data",
		Key:         "sensitive_data_exposure",
		Label:       "Sensitive data exposure",
		Description: "Sensitive or regulated data is flowing in telemetry where it should not.",
	},
}

func DescribeDomain(domain Domain) DomainDescriptor {
	if desc, ok := domainDescriptors[domain]; ok {
		return desc
	}
	label := strings.TrimSpace(string(domain))
	if label == "" {
		label = "Unknown"
	}
	return DomainDescriptor{
		Key:         domain,
		Label:       strings.Title(strings.ReplaceAll(label, "_", " ")),
		Description: "Curated findings derived from the durable fact substrate.",
	}
}

func DescribeType(domain Domain, key string) TypeDescriptor {
	if desc, ok := typeDescriptors[key]; ok {
		return desc
	}
	label := strings.TrimSpace(key)
	if label == "" {
		label = "unknown"
	}
	return TypeDescriptor{
		Domain:      domain,
		Dimension:   "",
		Key:         key,
		Label:       strings.Title(strings.ReplaceAll(label, "_", " ")),
		Description: "Curated finding type derived from the current production understanding layer.",
	}
}
