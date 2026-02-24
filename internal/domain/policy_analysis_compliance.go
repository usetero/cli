package domain

import (
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/format"
)

// Compliance category constants matching control plane schema.
const (
	CategoryPIILeakage         PolicyCategory = "pii_leakage"
	CategorySecretsLeakage     PolicyCategory = "secrets_leakage"
	CategoryPHILeakage         PolicyCategory = "phi_leakage"
	CategoryPaymentDataLeakage PolicyCategory = "payment_data_leakage"
)

// SensitiveField identifies a field that may contain sensitive data.
// Matches control plane's SensitiveField struct used across all compliance categories.
type SensitiveField struct {
	Path     FieldPath `json:"path"`     // Attribute path, e.g. ["attributes", "user", "email"]
	Types    []string  `json:"types"`    // Types of sensitive data this field could contain
	Observed bool      `json:"observed"` // True if actual sensitive data was seen in log values
}

// complianceSubtitle builds the one-line description shared by all compliance categories.
func complianceSubtitle(fields []SensitiveField) string {
	if len(fields) == 0 {
		return ""
	}
	f := fields[0]
	typesStr := strings.Join(f.Types, ", ")
	var s string
	if f.Observed {
		s = fmt.Sprintf("%s observed in %s", typesStr, f.Path.Key())
	} else {
		s = fmt.Sprintf("Potential %s in %s", typesStr, f.Path.Key())
	}
	if len(fields) > 1 {
		s += fmt.Sprintf(" (+%d more)", len(fields)-1)
	}
	return s
}

// complianceRelevantKeys builds the relevant attribute keys for all compliance categories.
func complianceRelevantKeys(fields []SensitiveField) []FieldPath {
	keys := make([]FieldPath, len(fields))
	for i, f := range fields {
		keys[i] = f.Path
	}
	return keys
}

// === PII Leakage ===

// PII type constants for display and filtering.
const (
	PIITypeEmail         = "email"
	PIITypeName          = "name"
	PIITypePhone         = "phone"
	PIITypeAddress       = "address"
	PIITypeSSN           = "ssn"
	PIITypeNationalID    = "national_id"
	PIITypeIPAddress     = "ip_address"
	PIITypeDateOfBirth   = "date_of_birth"
	PIITypeDriverLicense = "driver_license"
)

// PIILeakageAnalysis is the category-specific analysis for PII leakage policies.
type PIILeakageAnalysis struct {
	baseAnalysis
	Fields []SensitiveField `json:"fields"`
}

func (a PIILeakageAnalysis) Category() PolicyCategory  { return CategoryPIILeakage }
func (a PIILeakageAnalysis) Subtitle() string          { return complianceSubtitle(a.Fields) }
func (a PIILeakageAnalysis) RelevantKeys() []FieldPath { return complianceRelevantKeys(a.Fields) }

// === Secrets Leakage ===

// Secret type constants for display and filtering.
const (
	SecretTypeAPIKey             = "api_key"
	SecretTypeBearerToken        = "bearer_token"
	SecretTypeOAuthToken         = "oauth_token"
	SecretTypePassword           = "password"
	SecretTypePasswordHash       = "password_hash"
	SecretTypeDatabaseCredential = "database_credential"
	SecretTypeConnectionString   = "connection_string"
	SecretTypePrivateKey         = "private_key"
	SecretTypeCertificate        = "certificate"
	SecretTypeEncryptionKey      = "encryption_key"
	SecretTypeSigningKey         = "signing_key"
	SecretTypeWebhookSecret      = "webhook_secret"
	SecretTypeSessionToken       = "session_token"
)

// SecretsLeakageAnalysis is the category-specific analysis for secrets leakage policies.
type SecretsLeakageAnalysis struct {
	baseAnalysis
	Fields []SensitiveField `json:"fields"`
}

func (a SecretsLeakageAnalysis) Category() PolicyCategory { return CategorySecretsLeakage }
func (a SecretsLeakageAnalysis) Subtitle() string         { return complianceSubtitle(a.Fields) }
func (a SecretsLeakageAnalysis) RelevantKeys() []FieldPath {
	return complianceRelevantKeys(a.Fields)
}

// === PHI Leakage ===

// PHI type constants for display and filtering.
const (
	PHITypeDiagnosisCode       = "diagnosis_code"
	PHITypeProcedureCode       = "procedure_code"
	PHITypePrescription        = "prescription"
	PHITypeLabResult           = "lab_result"
	PHITypeMedicalRecordNumber = "medical_record_number"
	PHITypePatientIdentifier   = "patient_identifier"
	PHITypeHealthInsuranceID   = "health_insurance_id"
	PHITypeBiometric           = "biometric"
	PHITypeGeneticData         = "genetic_data"
)

// PHILeakageAnalysis is the category-specific analysis for PHI leakage policies.
type PHILeakageAnalysis struct {
	baseAnalysis
	Fields []SensitiveField `json:"fields"`
}

func (a PHILeakageAnalysis) Category() PolicyCategory  { return CategoryPHILeakage }
func (a PHILeakageAnalysis) Subtitle() string          { return complianceSubtitle(a.Fields) }
func (a PHILeakageAnalysis) RelevantKeys() []FieldPath { return complianceRelevantKeys(a.Fields) }

// === Payment Data Leakage ===

// Payment data type constants for display and filtering.
const (
	PaymentTypeCreditCard     = "credit_card"
	PaymentTypeCVV            = "cvv"
	PaymentTypePIN            = "pin"
	PaymentTypeBankAccount    = "bank_account"
	PaymentTypeRoutingNumber  = "routing_number"
	PaymentTypePaymentToken   = "payment_token"
	PaymentTypeMagneticStripe = "magnetic_stripe"
)

// PaymentDataLeakageAnalysis is the category-specific analysis for payment data leakage policies.
type PaymentDataLeakageAnalysis struct {
	baseAnalysis
	Fields []SensitiveField `json:"fields"`
}

func (a PaymentDataLeakageAnalysis) Category() PolicyCategory { return CategoryPaymentDataLeakage }
func (a PaymentDataLeakageAnalysis) Subtitle() string         { return complianceSubtitle(a.Fields) }
func (a PaymentDataLeakageAnalysis) RelevantKeys() []FieldPath {
	return complianceRelevantKeys(a.Fields)
}

// === Unified Compliance Policy ===

// ComplianceAnalysisEnvelope wraps the category-keyed analysis JSON from the database.
// Each field corresponds to one compliance category.
type ComplianceAnalysisEnvelope struct {
	PIILeakage         *PIILeakageAnalysis         `json:"pii_leakage,omitempty"`
	SecretsLeakage     *SecretsLeakageAnalysis     `json:"secrets_leakage,omitempty"`
	PHILeakage         *PHILeakageAnalysis         `json:"phi_leakage,omitempty"`
	PaymentDataLeakage *PaymentDataLeakageAnalysis `json:"payment_data_leakage,omitempty"`
}

// CompliancePolicy represents a single compliance finding (any category) with context from joined tables.
type CompliancePolicy struct {
	Category      PolicyCategory   // One of: pii_leakage, secrets_leakage, phi_leakage, payment_data_leakage
	LogEventName  string           // Log event name
	ServiceName   string           // Service name
	Fields        []SensitiveField // Parsed from analysis JSON
	VolumePerHour *float64         // Log event volume; nil when unmeasured
	AnyObserved   bool             // True if any field has observed sensitive data
}

// ComplianceCategorySummary provides counts and stats for a single compliance category.
type ComplianceCategorySummary struct {
	Category       PolicyCategory // One of the 4 compliance categories
	DisplayName    string         // Human-readable name from control plane
	Principle      string         // One-liner explaining what this category catches
	LeakingCount   int64          // Policies with observed sensitive data
	AtRiskCount    int64          // Policies without observed data (but flagged)
	FixedCount     int64          // Approved policies
	VolumePerHour  float64        // Total volume across all policies in this category
	ServiceCount   int            // Number of unique services affected
	UniqueServices []string
}

// Name returns the human-readable name for the compliance category.
// Uses the display name from the control plane, falling back to short labels.
func (c ComplianceCategorySummary) Name() string {
	if c.DisplayName != "" {
		return c.DisplayName
	}
	switch c.Category {
	case CategoryPIILeakage:
		return "PII"
	case CategorySecretsLeakage:
		return "Secrets"
	case CategoryPHILeakage:
		return "PHI"
	case CategoryPaymentDataLeakage:
		return "Payment Data"
	default:
		return format.TitleCase(string(c.Category))
	}
}
