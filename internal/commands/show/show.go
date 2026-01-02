package show

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mmdbasi/mtcli/internal/charts"
	"github.com/mmdbasi/mtcli/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <session_id>",
		Short: "Show details of a specific test session",
		Long: `Display detailed information about a specific typing test session.

Shows:
  - Full summary (WPM, raw WPM, accuracy, time)
  - Speed chart over the duration of the test
  - Mode and settings used`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(args[0])
		},
	}

	return cmd
}

func runShow(sessionIDStr string) error {
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid session ID: %s", sessionIDStr)
	}

	store, err := sqlite.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	session, err := store.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session %d not found", sessionID)
	}

	samples, err := store.GetSamples(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get samples: %w", err)
	}

	// Header
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Printf("  ║       SESSION #%-5d                 ║\n", session.ID)
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()

	// Session info
	fmt.Println("  Details")
	fmt.Println("  ────────────────────────────────────────")
	fmt.Printf("  Date:       %s\n", session.StartedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Mode:       %s\n", session.Mode)

	switch session.Mode {
	case "timer":
		fmt.Printf("  Duration:   %d seconds\n", session.Seconds)
	case "words":
		fmt.Printf("  Word count: %d words\n", session.Words)
	case "quote":
		if session.QuoteID != "" {
			fmt.Printf("  Quote ID:   %s\n", session.QuoteID)
		}
	}
	fmt.Println()

	// Results
	fmt.Println("  Results")
	fmt.Println("  ────────────────────────────────────────")
	fmt.Printf("  WPM:        %.1f\n", session.WPM)
	fmt.Printf("  Raw WPM:    %.1f\n", session.RawWPM)
	fmt.Printf("  Accuracy:   %.1f%%\n", session.Accuracy)
	fmt.Printf("  Time:       %s\n", formatDuration(time.Duration(session.DurationMs)*time.Millisecond))
	fmt.Printf("  Characters: %d/%d correct\n", session.CorrectChars, session.TotalTyped)
	fmt.Println()

	// Speed chart
	if len(samples) > 0 {
		fmt.Println("  Speed over time")
		fmt.Println("  ────────────────────────────────────────")
		fmt.Println()

		wpmPoints := make([]charts.DataPoint, len(samples))
		rawPoints := make([]charts.DataPoint, len(samples))
		for i, s := range samples {
			wpmPoints[i] = charts.DataPoint{TimeMs: s.TimeMs, Value: s.WPM}
			rawPoints[i] = charts.DataPoint{TimeMs: s.TimeMs, Value: s.RawWPM}
		}

		chartOpts := charts.DefaultOptions()
		chartOpts.Width = 60
		chartOpts.Height = 10
		chart := charts.RenderDualChart(wpmPoints, rawPoints, chartOpts)

		// Indent each line
		for _, line := range splitLines(chart) {
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Println()

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", m, s)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
