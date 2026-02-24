package compliance

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appmsg "github.com/usetero/cli/internal/app/msgs"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/components/table"
)

// detail renders the pending policies for a single compliance category.
type detail struct {
	theme    styles.Theme
	category domain.ComplianceCategorySummary
	policies []domain.CompliancePolicy
	cursor   int
}

// newDetail creates a detail view for the given category and pre-fetched policies.
func newDetail(theme styles.Theme, category domain.ComplianceCategorySummary, policies []domain.CompliancePolicy) *detail {
	return &detail{
		theme:    theme,
		category: category,
		policies: policies,
	}
}

// Prompt returns a tea.Cmd that emits a DrawerPrompt for the selected policy.
func (d *detail) Prompt() tea.Cmd {
	if len(d.policies) == 0 {
		return nil
	}
	p := d.policies[d.cursor]
	text := fmt.Sprintf(
		"Tell me about the %s compliance issue for the %q log event in the %s service.",
		d.category.Name(), p.LogEventName, p.ServiceName,
	)
	return func() tea.Msg { return appmsg.DrawerPrompt{Text: text} }
}

// View renders the detail: a header with category summary, then a policy table.
func (d *detail) View(width int) string {
	var lines []string
	lines = append(lines, d.renderHeader())
	if d.category.Principle != "" {
		lines = append(lines, "")
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		lines = append(lines, muted.Render(d.category.Principle))
	}
	lines = append(lines, "")

	if len(d.policies) == 0 {
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		lines = append(lines, muted.Render("No pending policies in this category."))
	} else {
		lines = append(lines, d.renderTable(width))
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the back hint + category name + summary.
func (d *detail) renderHeader() string {
	colors := d.theme
	muted := lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	text := lipgloss.NewStyle().Foreground(colors.Text).Background(colors.Bg)
	sep := muted.Render(" · ")

	back := muted.Render("esc ◀")
	name := text.Bold(true).Render(d.category.Name())

	var parts []string
	parts = append(parts, back+" "+name)

	if d.category.LeakingCount > 0 {
		err := lipgloss.NewStyle().Foreground(colors.Error).Background(colors.Bg)
		parts = append(parts, err.Render("●")+" "+err.Render(fmt.Sprintf("%d leaking", d.category.LeakingCount)))
	}
	if d.category.AtRiskCount > 0 {
		warn := lipgloss.NewStyle().Foreground(colors.Warning).Background(colors.Bg)
		parts = append(parts, warn.Render("●")+" "+muted.Render(fmt.Sprintf("%d at risk", d.category.AtRiskCount)))
	}

	if d.category.ServiceCount > 0 {
		parts = append(parts, muted.Render(fmt.Sprintf("%d services", d.category.ServiceCount)))
	}

	return strings.Join(parts, sep)
}

// renderTable renders the per-policy table.
func (d *detail) renderTable(width int) string {
	tbl := table.New(d.theme, table.WithMaxValueWidth(40))
	tbl.Headers("Log Event", "Service", "Volume", "Status")
	tbl.SetWidth(width)

	accent := lipgloss.NewStyle().Foreground(d.theme.Accent).Background(d.theme.Bg)

	for i, p := range d.policies {
		name := p.LogEventName
		if i == d.cursor {
			name = accent.Render("▶ " + name)
		} else {
			name = d.observedDot(p.AnyObserved) + " " + name
		}

		vol := "—"
		if p.VolumePerHour != nil {
			vol = format.Volume(*p.VolumePerHour) + " evt/hr"
		}

		status := d.formatSensitiveTypes(p.Fields, 3)

		tbl.Row(
			name,
			p.ServiceName,
			vol,
			status,
		)
	}

	return tbl.View()
}

// observedDot returns a colored dot based on whether sensitive data was observed.
// Red for observed (leaking), muted for at-risk.
func (d *detail) observedDot(observed bool) string {
	if observed {
		return lipgloss.NewStyle().Foreground(d.theme.Error).Background(d.theme.Bg).Render("●")
	}
	return lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg).Render("●")
}

// formatSensitiveTypes returns deduplicated type labels from all fields,
// showing at most maxShow before truncating with "+N".
func (d *detail) formatSensitiveTypes(fields []domain.SensitiveField, maxShow int) string {
	if len(fields) == 0 {
		return "—"
	}

	// Flatten all types across fields, deduplicate, preserve first-seen order.
	seen := make(map[string]struct{})
	var types []string
	for _, f := range fields {
		for _, t := range f.Types {
			label := displaySensitiveType(d.category.Category, t)
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{}
				types = append(types, label)
			}
		}
	}

	if len(types) == 0 {
		return "—"
	}

	visible := types
	remaining := 0
	if len(types) > maxShow {
		visible = types[:maxShow]
		remaining = len(types) - maxShow
	}

	result := strings.Join(visible, ", ")
	if remaining > 0 {
		muted := lipgloss.NewStyle().Foreground(d.theme.TextMuted).Background(d.theme.Bg)
		result += muted.Render(fmt.Sprintf(", +%d", remaining))
	}
	return result
}

// displaySensitiveType returns a human-readable label for a sensitive type value.
func displaySensitiveType(category domain.PolicyCategory, t string) string {
	// Category-specific type labels
	switch category {
	case domain.CategoryPIILeakage:
		return displayPIIType(t)
	case domain.CategorySecretsLeakage:
		return displaySecretType(t)
	case domain.CategoryPHILeakage:
		return displayPHIType(t)
	case domain.CategoryPaymentDataLeakage:
		return displayPaymentType(t)
	default:
		return t
	}
}

// displayPIIType returns a human-readable label for a PII type.
func displayPIIType(t string) string {
	switch t {
	case domain.PIITypeEmail:
		return "email"
	case domain.PIITypeName:
		return "name"
	case domain.PIITypePhone:
		return "phone"
	case domain.PIITypeAddress:
		return "address"
	case domain.PIITypeSSN:
		return "SSN"
	case domain.PIITypeNationalID:
		return "national ID"
	case domain.PIITypeIPAddress:
		return "IP address"
	case domain.PIITypeDateOfBirth:
		return "date of birth"
	case domain.PIITypeDriverLicense:
		return "driver license"
	default:
		return t
	}
}

// displaySecretType returns a human-readable label for a secret type.
func displaySecretType(t string) string {
	switch t {
	case domain.SecretTypeAPIKey:
		return "API key"
	case domain.SecretTypeBearerToken:
		return "bearer token"
	case domain.SecretTypeOAuthToken:
		return "OAuth token"
	case domain.SecretTypePassword:
		return "password"
	case domain.SecretTypePasswordHash:
		return "password hash"
	case domain.SecretTypeDatabaseCredential:
		return "database credential"
	case domain.SecretTypeConnectionString:
		return "connection string"
	case domain.SecretTypePrivateKey:
		return "private key"
	case domain.SecretTypeCertificate:
		return "certificate"
	case domain.SecretTypeEncryptionKey:
		return "encryption key"
	case domain.SecretTypeSigningKey:
		return "signing key"
	case domain.SecretTypeWebhookSecret:
		return "webhook secret"
	case domain.SecretTypeSessionToken:
		return "session token"
	default:
		return t
	}
}

// displayPHIType returns a human-readable label for a PHI type.
func displayPHIType(t string) string {
	switch t {
	case domain.PHITypeDiagnosisCode:
		return "diagnosis code"
	case domain.PHITypeProcedureCode:
		return "procedure code"
	case domain.PHITypePrescription:
		return "prescription"
	case domain.PHITypeLabResult:
		return "lab result"
	case domain.PHITypeMedicalRecordNumber:
		return "medical record number"
	case domain.PHITypePatientIdentifier:
		return "patient identifier"
	case domain.PHITypeHealthInsuranceID:
		return "health insurance ID"
	case domain.PHITypeBiometric:
		return "biometric"
	case domain.PHITypeGeneticData:
		return "genetic data"
	default:
		return t
	}
}

// displayPaymentType returns a human-readable label for a payment data type.
func displayPaymentType(t string) string {
	switch t {
	case domain.PaymentTypeCreditCard:
		return "credit card"
	case domain.PaymentTypeCVV:
		return "CVV"
	case domain.PaymentTypePIN:
		return "PIN"
	case domain.PaymentTypeBankAccount:
		return "bank account"
	case domain.PaymentTypeRoutingNumber:
		return "routing number"
	case domain.PaymentTypePaymentToken:
		return "payment token"
	case domain.PaymentTypeMagneticStripe:
		return "magnetic stripe"
	default:
		return t
	}
}
