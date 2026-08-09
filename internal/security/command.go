package security

import "strings"

type CommandInfo struct {
	Raw        string
	Executable string
	Arguments  []string

	HasPipe        bool
	HasRedirect    bool
	HasChain       bool
	HasShell       bool
	HasSudo        bool
	HasNetwork     bool
	HasDestructive bool
}

var protectedPaths = []string{
	".env",
	".env.",
	".git/",
	"internal/security/",
}

func touchesProtectedPath(command string) bool {
	lower := strings.ToLower(command)

	for _, path := range protectedPaths {
		if strings.Contains(lower, path) {
			return true
		}
	}

	return false
}

func AnalyzeCommand(command string) CommandInfo {
	info := CommandInfo{
		Raw: command,
	}

	parts := strings.Fields(command)

	if len(parts) > 0 {
		info.Executable = parts[0]
		info.Arguments = parts[1:]
	}

	lower := strings.ToLower(command)

	info.HasPipe = strings.Contains(command, "|")

	info.HasRedirect = strings.ContainsAny(
		command,
		"><",
	)

	info.HasChain =
		strings.Contains(command, "&&") ||
			strings.Contains(command, "||") ||
			strings.Contains(command, ";")

	info.HasShell =
		strings.Contains(lower, "sh -c") ||
			strings.Contains(lower, "bash -c") ||
			strings.Contains(lower, "zsh -c")

	info.HasSudo =
		info.Executable == "sudo" ||
			info.Executable == "doas" ||
			info.Executable == "su"

	info.HasNetwork = isNetworkCommand(
		info.Executable,
	)

	info.HasDestructive = isDestructiveCommand(
		lower,
	)

	return info
}

func isNetworkCommand(executable string) bool {
	switch executable {
	case
		"curl",
		"wget",
		"nc",
		"netcat",
		"ssh",
		"scp",
		"rsync",
		"git":
		return true
	default:
		return false
	}
}

func isDestructiveCommand(command string) bool {
	dangerous := []string{
		"rm -rf /",
		"rm -rf /*",
		"rm -r /",
		"mkfs",
		"fdisk",
		"parted",
		"dd if=",
		"shutdown",
		"reboot",
		"poweroff",
		"halt",
		":(){",
	}

	for _, pattern := range dangerous {
		if strings.Contains(command, pattern) {
			return true
		}
	}

	return false
}
