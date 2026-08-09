package security

type Decision int

const (
	DecisionAllow Decision = iota
	DecisionApproval
	DecisionBlock
)

type Policy struct {
	MaxCommandLength int
	MaxOutputSize    int

	AllowShell          bool
	AllowSudo           bool
	AllowNetwork        bool
	AllowPackageInstall bool
}

func DefaultPolicy() Policy {
	return Policy{
		MaxCommandLength: 4096,
		MaxOutputSize:    32 * 1024,

		AllowShell:          false,
		AllowSudo:           false,
		AllowNetwork:        true,
		AllowPackageInstall: false,
	}
}
