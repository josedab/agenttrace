package service

import (
	"fmt"
	"strings"

	"github.com/agenttrace/agenttrace/api/internal/domain"
)

// BuildOutcomeDigest creates channel-neutral report content without inventing unavailable metrics.
func BuildOutcomeDigest(overview *domain.OutcomeOverview) *domain.OutcomeDigest {
	digest := &domain.OutcomeDigest{
		ProjectID:   overview.ProjectID,
		Period:      overview.Period,
		Title:       "Agent outcome digest",
		Highlights:  []string{},
		Attention:   []string{},
		GeneratedAt: overview.GeneratedAt,
	}

	if overview.Runs.SuccessRate.Available {
		digest.Summary = fmt.Sprintf(
			"%d agent runs completed with a %.1f%% success rate.",
			overview.Runs.Total,
			*overview.Runs.SuccessRate.Value*100,
		)
		digest.Highlights = append(
			digest.Highlights,
			fmt.Sprintf("%d successful agent outcomes", overview.Runs.Successful),
		)
	} else {
		digest.Summary = "No completed agent outcomes were available for this period."
	}

	if overview.CI.PassRate.Available {
		digest.Highlights = append(
			digest.Highlights,
			fmt.Sprintf(
				"CI passed %.1f%% of %d runs",
				*overview.CI.PassRate.Value*100,
				overview.CI.Total,
			),
		)
	} else {
		digest.Attention = append(digest.Attention, "CI outcome data is unavailable")
	}

	if overview.Cost.CostPerSuccessfulOutcome.Available {
		digest.Highlights = append(
			digest.Highlights,
			fmt.Sprintf(
				"Cost per successful outcome: $%.4f",
				*overview.Cost.CostPerSuccessfulOutcome.Value,
			),
		)
	}
	if overview.SourceControl.RegressionSignals > 0 {
		digest.Attention = append(
			digest.Attention,
			fmt.Sprintf(
				"%d linked commit(s) had failing CI signals",
				overview.SourceControl.RegressionSignals,
			),
		)
	}
	if overview.SourceControl.RevertSignals > 0 {
		digest.Attention = append(
			digest.Attention,
			fmt.Sprintf(
				"%d revert commit signal(s) were detected",
				overview.SourceControl.RevertSignals,
			),
		)
	}
	if len(overview.Availability.Unavailable) > 0 {
		digest.Attention = append(
			digest.Attention,
			"Unavailable: "+strings.Join(overview.Availability.Unavailable, ", "),
		)
	}

	return digest
}

// RenderOutcomeDigestMarkdown renders a digest for GitHub, Slack, Discord, or generic webhooks.
func RenderOutcomeDigestMarkdown(digest *domain.OutcomeDigest) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "## %s\n\n%s\n", digest.Title, digest.Summary)
	fmt.Fprintf(
		&builder,
		"\nPeriod: %s to %s\n",
		digest.Period.From.Format("2006-01-02"),
		digest.Period.To.Format("2006-01-02"),
	)

	if len(digest.Highlights) > 0 {
		builder.WriteString("\n### Highlights\n")
		for _, item := range digest.Highlights {
			fmt.Fprintf(&builder, "- %s\n", item)
		}
	}
	if len(digest.Attention) > 0 {
		builder.WriteString("\n### Needs attention\n")
		for _, item := range digest.Attention {
			fmt.Fprintf(&builder, "- %s\n", item)
		}
	}
	return strings.TrimSpace(builder.String())
}
