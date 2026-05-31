package risk

type UncertaintyLevel string

const (
	UncertaintyLow      UncertaintyLevel = "LOW"
	UncertaintyMedium   UncertaintyLevel = "MEDIUM"
	UncertaintyHigh     UncertaintyLevel = "HIGH"
	UncertaintyVeryHigh UncertaintyLevel = "VERY HIGH"
)

type UncertaintyScore struct {
	Level   UncertaintyLevel
	Reasons []string
	Score   float64
}

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type RiskScore struct {
	Score        float64
	Level        RiskLevel
	Contributors []Contributor
}

type Contributor struct {
	RuleID       string
	Name         string
	Weight       int
	Value        float64
	Contribution float64
	Details      string
}
