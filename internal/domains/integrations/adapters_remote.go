package integrations

import controlplane "github.com/usetero/cli/internal/infrastructure/controlplane/api"

func toControlPlaneDatadogAccountID(id DatadogAccountID) controlplane.DatadogAccountID {
	return controlplane.DatadogAccountID(id)
}

func toControlPlaneDatadogSite(site DatadogSite) controlplane.DatadogSite {
	return controlplane.DatadogSite(site)
}

func toControlPlaneCreateDatadogAccountInput(input DatadogAccountCreate) controlplane.CreateDatadogAccountInput {
	return controlplane.CreateDatadogAccountInput{
		Name:   input.Name.String(),
		Site:   toControlPlaneDatadogSite(input.Site),
		APIKey: input.APIKey.String(),
		AppKey: input.AppKey.String(),
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
		Health:                        DatadogAccountHealth(in.Health),
		ReadyForUse:                   in.ReadyForUse,
		ServiceCount:                  in.ServiceCount,
		ActiveServices:                in.ActiveServices,
		OKServices:                    in.OKServices,
		DisabledServices:              in.DisabledServices,
		InactiveServices:              in.InactiveServices,
		EventCount:                    in.EventCount,
		AnalyzedCount:                 in.AnalyzedCount,
		PreviewLogEventCount:          in.PreviewLogEventCount,
		EffectiveLogEventCount:        in.EffectiveLogEventCount,
		CurrentEventsPerHour:          in.CurrentEventsPerHour,
		CurrentBytesPerHour:           in.CurrentBytesPerHour,
		CurrentTotalUSDPerHour:        in.CurrentTotalUSDPerHour,
		PreviewSavedEventsPerHour:     in.PreviewSavedEventsPerHour,
		PreviewSavedBytesPerHour:      in.PreviewSavedBytesPerHour,
		PreviewSavedTotalUSDPerHour:   in.PreviewSavedTotalUSDPerHour,
		EffectiveSavedEventsPerHour:   in.EffectiveSavedEventsPerHour,
		EffectiveSavedBytesPerHour:    in.EffectiveSavedBytesPerHour,
		EffectiveSavedTotalUSDPerHour: in.EffectiveSavedTotalUSDPerHour,
		RefreshedAt:                   in.RefreshedAt.UTC(),
	}
}
