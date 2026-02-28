package onboarding

import appmsg "github.com/usetero/cli/internal/app/onboarding/msgs"

func defaultStatusForGate(gate Gate) appmsg.StepStatus {
	switch gate {
	case GatePreflight:
		return appmsg.StepStatus{Title: "Getting ready", Details: "Preparing onboarding..."}
	case GateAuthenticate:
		return appmsg.StepStatus{Title: "Sign in", Details: "Waiting for authentication..."}
	case GateRoleSelect:
		return appmsg.StepStatus{Title: "Choose role", Details: "Select your onboarding role."}
	case GateOrgSelect:
		return appmsg.StepStatus{Title: "Choose organization", Details: "Select an organization to continue."}
	case GateOrgCreate:
		return appmsg.StepStatus{Title: "Create organization", Details: "Create an organization to continue."}
	case GateAccountSelect:
		return appmsg.StepStatus{Title: "Choose account", Details: "Select an account to continue."}
	case GateAccountCreate:
		return appmsg.StepStatus{Title: "Create account", Details: "Create an account to continue."}
	case GateRuntimeInit:
		return appmsg.StepStatus{Title: "Getting ready", Details: "Initializing account runtime..."}
	case GateDatadogCheck:
		return appmsg.StepStatus{Title: "Getting ready", Details: "Checking Datadog configuration..."}
	case GateDatadogRegion:
		return appmsg.StepStatus{Title: "Datadog setup", Details: "Choose your Datadog site."}
	case GateDatadogAPIKey:
		return appmsg.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog API key."}
	case GateDatadogAppKey:
		return appmsg.StepStatus{Title: "Datadog setup", Details: "Enter your Datadog application key."}
	case GateDatadogDiscovery:
		return appmsg.StepStatus{Title: "Datadog setup", Details: "Discovering Datadog services..."}
	case GateWorkspaceSelect:
		return appmsg.StepStatus{Title: "Choose workspace", Details: "Loading workspaces..."}
	case GateSync:
		return appmsg.StepStatus{Title: "Getting ready", Details: "Syncing your data..."}
	default:
		return appmsg.StepStatus{Title: "Getting ready", Details: "Preparing onboarding..."}
	}
}
