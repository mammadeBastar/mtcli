package history

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmdbasi/mtcli/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

// Options holds the history command options
type Options struct {
	Limit int
	Mode  string
}

func NewHistoryCmd() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show your test history",
		Long: `Display a list of your recent typing tests.

Shows date, mode, WPM, raw WPM, accuracy, duration, and session ID for each test.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Limit, "limit", "n", 20, "number of sessions to show")
	cmd.Flags().StringVarP(&opts.Mode, "mode", "m", "", "filter by mode (timer, words, quote)")

	return cmd
}

func runHistory(opts *Options) error {
	store, err := sqlite.Open()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	sessions, err := store.ListSessions(opts.Limit, opts.Mode)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("\n  No typing tests recorded yet.")
		if opts.Mode != "" {
			fmt.Printf("  (filtered by mode: %s)\n", opts.Mode)
		}
		fmt.Println("  Run 'mtcli test' to start your first test!")
		fmt.Println()
		return nil
	}

	// Header
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                         TEST HISTORY                                 ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Table header
	fmt.Println("  ID    Date                 Mode    WPM     Raw     Acc      Time")
	fmt.Println("  ────────────────────────────────────────────────────────────────────────")

	for _, session := range sessions {
		// Format date
		dateStr := session.StartedAt.Format("2006-01-02 15:04")

		// Format mode with fixed width
		modeStr := padRight(session.Mode, 6)

		// Format duration
		durationStr := formatDuration(time.Duration(session.DurationMs) * time.Millisecond)

		fmt.Printf("  %-5d %s  %s  %5.1f   %5.1f   %5.1f%%  %s\n",
			session.ID,
			dateStr,
			modeStr,
			session.WPM,
			session.RawWPM,
			session.Accuracy,
			durationStr,
		)
	}

	fmt.Println()
	fmt.Printf("  Showing %d most recent tests", len(sessions))
	if opts.Mode != "" {
		fmt.Printf(" (mode: %s)", opts.Mode)
	}
	fmt.Println()
	fmt.Println("  Use 'mtcli show <id>' to see details of a specific test.")
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

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}
