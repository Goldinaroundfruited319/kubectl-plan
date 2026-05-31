package output

import (
	"fmt"
	"math"
	"strings"

	"github.com/fatih/color"
	"github.com/samaasi/kubectl-plan/internal/analysis"
	"github.com/samaasi/kubectl-plan/internal/dependency"
	"github.com/samaasi/kubectl-plan/internal/risk"
)

type DoctorResult struct {
	Namespace           string
	K8sAPIReachable     bool
	EstimatedConfidence float64
}

func (r *Renderer) RenderTerminal(res *analysis.AnalysisResult) error {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Fprintf(r.writer, "%s     %s %s  [%s: %s]\n\n",
		bold.Sprint("ACTION:"),
		cyan.Sprint(res.Action),
		cyan.Sprint(res.Target.Kind+"/"+res.Target.Name),
		bold.Sprint("namespace"),
		res.Target.Namespace,
	)

	riskColor := color.New(color.FgGreen)
	switch res.Risk.Level {
	case risk.RiskMedium:
		riskColor = color.New(color.FgYellow)
	case risk.RiskHigh:
		riskColor = color.New(color.FgRed)
	case risk.RiskCritical:
		riskColor = color.New(color.FgRed, color.Bold)
	}

	riskBar := makeProgressBar(res.Risk.Score/10.0, r.ascii)
	fmt.Fprintf(r.writer, "%s       %s / 10  %s  %s\n",
		bold.Sprint("RISK SCORE:"),
		riskColor.Sprintf("%.1f", res.Risk.Score),
		riskBar,
		riskColor.Sprint(res.Risk.Level),
	)

	confColor := color.New(color.FgYellow)
	if res.Confidence.Overall >= 0.85 {
		confColor = color.New(color.FgGreen)
	} else if res.Confidence.Overall < 0.65 {
		confColor = color.New(color.FgRed)
	}
	confBar := makeProgressBar(res.Confidence.Overall, r.ascii)
	fmt.Fprintf(r.writer, "%s        %s      %s  (%s)\n",
		bold.Sprint("CONFIDENCE:"),
		confColor.Sprintf("%d%%", int(res.Confidence.Overall*100)),
		confBar,
		strings.Join(res.Confidence.Sources, " + "),
	)

	uncColor := color.New(color.FgGreen)
	switch res.Uncertainty.Level {
	case risk.UncertaintyMedium:
		uncColor = color.New(color.FgYellow)
	case risk.UncertaintyHigh:
		uncColor = color.New(color.FgRed)
	case risk.UncertaintyVeryHigh:
		uncColor = color.New(color.FgRed, color.Bold)
	}
	fmt.Fprintf(r.writer, "%s       %s      (%s)\n\n",
		bold.Sprint("UNCERTAINTY:"),
		uncColor.Sprint(res.Uncertainty.Level),
		strings.Join(res.Uncertainty.Reasons, " · "),
	)

	fmt.Fprintln(r.writer, bold.Sprint("DEPENDENTS:"))

	var rootEdges []dependency.Edge
	for _, edge := range res.Graph.Edges {
		if edge.Relationship != dependency.RelOwns {
			rootEdges = append(rootEdges, edge)
		}
	}

	if len(rootEdges) == 0 {
		fmt.Fprintln(r.writer, "  No dependents detected.")
	} else {
		tLine := "├─"
		bLine := "└─"
		vLine := "│ "
		if r.ascii {
			tLine = "|-"
			bLine = "`-"
			vLine = "| "
		}

		for i, edge := range rootEdges {
			prefix := tLine
			if i == len(rootEdges)-1 {
				prefix = bLine
			}

			fromParts := strings.Split(edge.From, "/")
			fromName := fromParts[len(fromParts)-1]

			isLowConf := edge.Confidence < 0.65
			tilde := ""
			if isLowConf {
				tilde = "~"
			}

			directStr := "DIRECT"
			if edge.Depth > 1 {
				directStr = "INDIRECT"
			}

			fmt.Fprintf(r.writer, "  %s %s%s   %-8s  [%d%%]\n",
				prefix,
				tilde,
				fromName,
				directStr,
				int(edge.Confidence*100),
			)

			nextPrefix := "  " + vLine
			if i == len(rootEdges)-1 {
				nextPrefix = "    "
			}

			for _, ev := range edge.Evidence {
				fmt.Fprintf(r.writer, "%s     Evidence: %s\n",
					nextPrefix,
					ev.Description,
				)
			}
			if i < len(rootEdges)-1 {
				fmt.Fprintf(r.writer, "%s\n", nextPrefix)
			}
		}
	}
	fmt.Fprintln(r.writer)

	fmt.Fprintln(r.writer, bold.Sprint("UNKNOWN BLAST RADIUS:"))
	fmt.Fprintln(r.writer, "  ⚠ Cannot detect: Kafka consumers, external HTTP clients, Consul-registered services")
	fmt.Fprintln(r.writer, "  ℹ Run `kubectl plan doctor` to see what instrumentation would increase confidence.")
	fmt.Fprintln(r.writer)

	fmt.Fprintln(r.writer, bold.Sprint("RISK CONTRIBUTORS:"))
	for _, contrib := range res.Risk.Contributors {
		if contrib.Value > 0 {
			fmt.Fprintf(r.writer, "  +%.1f  %-30s  [%s]\n",
				contrib.Contribution,
				contrib.Name,
				contrib.Details,
			)
		}
	}
	fmt.Fprintf(r.writer, "  -----\n  = %.1f / 10\n\n", res.Risk.Score)

	fmt.Fprintln(r.writer, bold.Sprint("RECOMMENDATION:"))
	if res.Risk.Score >= 8.6 {
		fmt.Fprintln(r.writer, "  ⚠ CRITICAL RISK: Highly recommend holding off or carrying out during off-peak windows.")
	} else if res.Risk.Score >= 6.1 {
		fmt.Fprintln(r.writer, "  ⚠ High risk operation. Proceed with caution and ensure rollback plans are active.")
	} else {
		fmt.Fprintln(r.writer, "  ✓ Low risk operation. Safe to proceed.")
	}

	return nil
}

func (r *Renderer) RenderWhy(res *analysis.AnalysisResult) error {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Fprintf(r.writer, "%s %s\n\n",
		bold.Sprint("RISK SCORE BREAKDOWN:"),
		cyan.Sprint(res.Target.Kind+"/"+res.Target.Name),
	)

	riskColor := color.New(color.FgGreen)
	switch res.Risk.Level {
	case risk.RiskMedium:
		riskColor = color.New(color.FgYellow)
	case risk.RiskHigh:
		riskColor = color.New(color.FgRed)
	case risk.RiskCritical:
		riskColor = color.New(color.FgRed, color.Bold)
	}

	riskBar := makeProgressBar(res.Risk.Score/10.0, r.ascii)
	fmt.Fprintf(r.writer, "%-12s %.1f / 10  %s  %s\n",
		bold.Sprint("Score:"),
		res.Risk.Score,
		riskBar,
		riskColor.Sprint(res.Risk.Level),
	)

	confColor := color.New(color.FgYellow)
	if res.Confidence.Overall >= 0.85 {
		confColor = color.New(color.FgGreen)
	} else if res.Confidence.Overall < 0.65 {
		confColor = color.New(color.FgRed)
	}
	confBar := makeProgressBar(res.Confidence.Overall, r.ascii)
	fmt.Fprintf(r.writer, "%-12s %-12s %s  (%s)\n",
		bold.Sprint("Confidence:"),
		confColor.Sprintf("%d%%", int(res.Confidence.Overall*100)),
		confBar,
		strings.Join(res.Confidence.Sources, " + "),
	)

	uncColor := color.New(color.FgGreen)
	switch res.Uncertainty.Level {
	case risk.UncertaintyMedium:
		uncColor = color.New(color.FgYellow)
	case risk.UncertaintyHigh:
		uncColor = color.New(color.FgRed)
	case risk.UncertaintyVeryHigh:
		uncColor = color.New(color.FgRed, color.Bold)
	}
	fmt.Fprintf(r.writer, "%-12s %s         (%s)\n\n",
		bold.Sprint("Uncertainty:"),
		uncColor.Sprint(res.Uncertainty.Level),
		strings.Join(res.Uncertainty.Reasons, " · "),
	)

	fmt.Fprintln(r.writer, bold.Sprint("CONTRIBUTORS:"))
	fmt.Fprintf(r.writer, "  %-32s %-8s %-7s %s\n", "Rule", "Weight", "Value", "Contribution")
	fmt.Fprintln(r.writer, "  "+strings.Repeat("─", 65))

	for _, contrib := range res.Risk.Contributors {
		fmt.Fprintf(r.writer, "  %-32s x%-7d %-7.2f +%.1f\n",
			contrib.Name,
			contrib.Weight,
			contrib.Value,
			contrib.Contribution,
		)
	}
	fmt.Fprintln(r.writer, "  "+strings.Repeat("─", 65))
	fmt.Fprintf(r.writer, "  Total                                             %.1f / 10\n\n", res.Risk.Score)

	fmt.Fprintln(r.writer, bold.Sprint("CONFIDENCE SOURCES:"))
	fmt.Fprintln(r.writer, "  ✓ Kubernetes topology    (label selectors, ingress routing, owner references)")
	fmt.Fprintln(r.writer, "  ? Prometheus traffic     (Prometheus integration is inactive/not available)")
	fmt.Fprintln(r.writer)

	fmt.Fprintln(r.writer, bold.Sprint("UNKNOWN BLAST RADIUS:"))
	fmt.Fprintln(r.writer, "  ⚠ Kafka consumers, external HTTP clients, Consul-registered services")

	return nil
}

func (r *Renderer) RenderDoctor(res *DoctorResult) error {
	bold := color.New(color.Bold)

	fmt.Fprintf(r.writer, "%s  [namespace: %s]\n\n",
		bold.Sprint("ANALYSIS READINESS"),
		res.Namespace,
	)

	fmt.Fprintln(r.writer, bold.Sprint("DATA SOURCES:"))
	if res.K8sAPIReachable {
		fmt.Fprintln(r.writer, "  ✓ Kubernetes API          reachable · resources scanned successfully")
	} else {
		fmt.Fprintln(r.writer, "  ✗ Kubernetes API          unreachable")
	}

	fmt.Fprintln(r.writer, "  ✗ Prometheus              not integrated/configured [Week 1 MVP]")
	fmt.Fprintln(r.writer, "                            Confidence boost: +0%")
	fmt.Fprintln(r.writer, "  ✗ Istio / Service Mesh    not detected")
	fmt.Fprintln(r.writer, "  ✗ OpenTelemetry           not detected")
	fmt.Fprintln(r.writer, "  ✗ Historical records      no records store found")
	fmt.Fprintln(r.writer)

	fmt.Fprintln(r.writer, bold.Sprint("NAMESPACE CRITICALITY PROFILE:"))
	fmt.Fprintln(r.writer, "  ✓ Default fallback active: namespace matching 'prod' -> MEDIUM risk (0.60)")
	fmt.Fprintln(r.writer, "  (Create ~/.kubectl-plan/criticality.yaml to define custom criticality profiles)")
	fmt.Fprintln(r.writer)

	confBar := makeProgressBar(res.EstimatedConfidence, r.ascii)
	fmt.Fprintf(r.writer, "%s\n  %d%%  %s\n\n",
		bold.Sprint("ESTIMATED ANALYSIS CONFIDENCE:"),
		int(res.EstimatedConfidence*100),
		confBar,
	)

	fmt.Fprintln(r.writer, bold.Sprint("TO IMPROVE CONFIDENCE:"))
	fmt.Fprintln(r.writer, "  → Integrate Prometheus data source (v0.2)")
	fmt.Fprintln(r.writer, "  → Install Istio or Linkerd for traffic topology evidence (v0.3)")
	fmt.Fprintln(r.writer, "  → Create historical record store (v0.4)")

	return nil
}

func makeProgressBar(pct float64, ascii bool) string {
	pct = math.Max(0.0, math.Min(1.0, pct))
	filled := int(math.Round(pct * 10))
	if filled > 10 {
		filled = 10
	}
	empty := 10 - filled

	charFilled := "█"
	charEmpty := "░"
	if ascii {
		charFilled = "#"
		charEmpty = "-"
	}

	return strings.Repeat(charFilled, filled) + strings.Repeat(charEmpty, empty)
}
