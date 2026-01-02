package stats

import (
	"fmt"
	"time"

	"github.com/mmdbasi/mtcli/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func NewStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show your typing statistics",
		Long: `Display aggregate statistics from all your typing tests.

Shows:
  - Total tests completed
  - Average WPM and best WPM
  - Average accuracy
  - Recent trends (last 7/30 days)
  - Breakdown by mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats()
		},
	}

	return cmd
}

func runStats() error {
	store, err := sqlite.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	stats, err := store.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	if stats.TotalTests == 0 {
		fmt.Println("\n  No typing tests recorded yet.")
		fmt.Println("  Run 'mtcli test' to start your first test!")
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       YOUR TYPING STATISTICS         ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()

	// Overall stats
	fmt.Println("  Overall")
	fmt.Println("  ────────────────────────────────────────")
	fmt.Printf("  Total Tests:      %d\n", stats.TotalTests)
	fmt.Printf("  Total Time:       %s\n", formatDuration(time.Duration(stats.TotalTimeMs)*time.Millisecond))
	fmt.Printf("  Average WPM:      %.1f\n", stats.AverageWPM)
	fmt.Printf("  Best WPM:         %.1f\n", stats.BestWPM)
	fmt.Printf("  Average Accuracy: %.1f%%\n", stats.AverageAccuracy)
	fmt.Println()

	// Recent trends
	fmt.Println("  Recent Trends")
	fmt.Println("  ────────────────────────────────────────")
	fmt.Printf("  Last 7 days avg:  %.1f WPM\n", stats.Last7DaysAvgWPM)
	fmt.Printf("  Last 30 days avg: %.1f WPM\n", stats.Last30DaysAvgWPM)

	// Trend indicator
	if stats.Last7DaysAvgWPM > 0 && stats.Last30DaysAvgWPM > 0 {
		diff := stats.Last7DaysAvgWPM - stats.Last30DaysAvgWPM
		if diff > 2 {
			fmt.Printf("  Trend:            ↑ Improving (+%.1f WPM)\n", diff)
		} else if diff < -2 {
			fmt.Printf("  Trend:            ↓ Declining (%.1f WPM)\n", diff)
		} else {
			fmt.Println("  Trend:            → Stable")
		}
	}
	fmt.Println()

	// Per-mode breakdown
	if len(stats.ModeStats) > 0 {
		fmt.Println("  By Mode")
		fmt.Println("  ────────────────────────────────────────")
		for mode, modeStats := range stats.ModeStats {
			fmt.Printf("  %s:\n", mode)
			fmt.Printf("    Tests: %d | Avg: %.1f WPM | Best: %.1f WPM\n",
				modeStats.TestCount, modeStats.AverageWPM, modeStats.BestWPM)
		}
		fmt.Println()
	}

	return nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
