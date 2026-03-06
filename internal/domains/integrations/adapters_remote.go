package integrations

import (
	"github.com/usetero/cli/internal/domains/tenancy"
	controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"
)

func toControlPlaneAccountID(id tenancy.AccountID) controlplane.AccountID {
	return controlplane.AccountID(id)
}

func toControlPlaneDatadogAccountID(id DatadogAccountID) controlplane.DatadogAccountID {
	return controlplane.DatadogAccountID(id)
}

func toControlPlaneDatadogSite(site DatadogSite) controlplane.DatadogSite {
	return controlplane.DatadogSite(site)
}

func toControlPlaneCreateDatadogAccountInput(input DatadogAccountCreate) controlplane.CreateDatadogAccountInput {
	return controlplane.CreateDatadogAccountInput{
		AccountID: toControlPlaneAccountID(input.AccountID),
		Name:      input.Name.String(),
		Site:      toControlPlaneDatadogSite(input.Site),
		APIKey:    input.APIKey.String(),
		AppKey:    input.AppKey.String(),
	}
}

func fromControlPlaneDatadogAccount(in controlplane.DatadogAccount) DatadogAccount {
	return DatadogAccount{
		ID:   DatadogAccountID(in.ID),
		Name: in.Name,
		Site: DatadogSite(in.Site),
	}
}

func fromControlPlaneDatadogStatus(in controlplane.DatadogAccountStatus) DatadogStatus {
	return DatadogStatus{
		Health:               DatadogAccountHealth(in.Health),
		ReadyForUse:          in.ReadyForUse,
		ServiceCount:         in.ServiceCount,
		ActiveServices:       in.ActiveServices,
		OKServices:           in.OKServices,
		DisabledServices:     in.DisabledServices,
		InactiveServices:     in.InactiveServices,
		EventCount:           in.EventCount,
		AnalyzedCount:        in.AnalyzedCount,
		PendingPolicyCount:   in.PendingPolicyCount,
		ApprovedPolicyCount:  in.ApprovedPolicyCount,
		DismissedPolicyCount: in.DismissedPolicyCount,
	}
}
