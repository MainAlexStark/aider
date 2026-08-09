package agent

type Risk string

const (
	RiskLow      Risk = "low"
	RiskMedium   Risk = "medium"
	RiskHigh     Risk = "high"
	RiskCritical Risk = "critical"
)

type Action struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Risk    Risk   `json:"risk"`
}

type Solution struct {
	Problem     string   `json:"problem"`
	Explanation string   `json:"explanation"`
	Confidence  float64  `json:"confidence"`
	Risk        Risk     `json:"risk"`
	Actions     []Action `json:"actions"`
}
