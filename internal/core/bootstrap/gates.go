package bootstrap

// Gate represents the next bootstrap stage to run.
type Gate string

const (
	GateAuthenticate  Gate = "authenticate"
	GateRoleSelect    Gate = "role_select"
	GateOrgSelect     Gate = "org_select"
	GateAccountSelect Gate = "account_select"
	GateRuntimeInit   Gate = "runtime_init"
)
