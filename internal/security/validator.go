package security

import (
	"fmt"
	"strings"
)

func Validate(
	command string,
	policy Policy,
) (Decision, error) {

	command = strings.TrimSpace(command)

	if command == "" {
		return DecisionBlock, fmt.Errorf(
			"empty command",
		)
	}

	if len(command) > policy.MaxCommandLength {
		return DecisionBlock, fmt.Errorf(
			"command exceeds maximum length: %d",
			policy.MaxCommandLength,
		)
	}

	if touchesProtectedPath(command) {
		return DecisionBlock, fmt.Errorf(
			"command accesses a protected path",
		)
	}

	info := AnalyzeCommand(command)

	if info.HasDestructive {
		return DecisionBlock, fmt.Errorf(
			"destructive command is blocked",
		)
	}

	if info.HasSudo && !policy.AllowSudo {
		return DecisionBlock, fmt.Errorf(
			"privileged commands are disabled",
		)
	}

	if info.HasShell && !policy.AllowShell {
		return DecisionBlock, fmt.Errorf(
			"shell execution is disabled",
		)
	}

	if info.HasPipe {
		return DecisionApproval, fmt.Errorf(
			"command contains a pipe",
		)
	}

	if info.HasRedirect {
		return DecisionApproval, fmt.Errorf(
			"command contains shell redirection",
		)
	}

	if info.HasChain {
		return DecisionApproval, fmt.Errorf(
			"command contains chained commands",
		)
	}

	if info.HasNetwork && !policy.AllowNetwork {
		return DecisionBlock, fmt.Errorf(
			"network access is disabled",
		)
	}

	return DecisionAllow, nil
}
