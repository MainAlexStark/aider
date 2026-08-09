package security

import "regexp"

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(
		`(?i)(api[_-]?key\s*[=:]\s*)[^\s]+`,
	),

	regexp.MustCompile(
		`(?i)(token\s*[=:]\s*)[^\s]+`,
	),

	regexp.MustCompile(
		`(?i)(password\s*[=:]\s*)[^\s]+`,
	),

	regexp.MustCompile(
		`(?i)(secret\s*[=:]\s*)[^\s]+`,
	),

	regexp.MustCompile(
		`(?i)(authorization:\s*bearer\s+)[^\s]+`,
	),
}

func Redact(text string) string {
	for _, pattern := range secretPatterns {
		text = pattern.ReplaceAllString(
			text,
			"${1}[REDACTED]",
		)
	}

	return text
}

func LimitOutput(
	text string,
	maxSize int,
) string {

	if len(text) <= maxSize {
		return text
	}

	return text[:maxSize] +
		"\n...[OUTPUT TRUNCATED]..."
}
