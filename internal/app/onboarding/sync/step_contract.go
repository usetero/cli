package sync

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func (m *Model) Hidden() bool { return false }

func (m *Model) Status() appmsg.StepStatus {
	return appmsg.StepStatus{Title: "Getting ready", Details: "Syncing your data..."}
}
