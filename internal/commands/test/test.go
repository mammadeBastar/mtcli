package test

import (
	"fmt"
	"time"

	"github.com/mmdbasi/mtcli/internal/charts"
	"github.com/mmdbasi/mtcli/internal/config"
	"github.com/mmdbasi/mtcli/internal/input"
	"github.com/mmdbasi/mtcli/internal/storage/sqlite"
	"github.com/mmdbasi/mtcli/internal/test"
	"github.com/mmdbasi/mtcli/internal/text"
	"github.com/mmdbasi/mtcli/internal/ui"
	"github.com/spf13/cobra"
)

// Options holds the test command options
type Options struct {
	Mode        string
	Seconds     int
	Words       int
	QuoteID     string
	QuoteRandom bool
	QuotesFile  string
	WordsFile   string
	Countdown   int
	Seed        int64
	NoColor     bool
	Wrap        int
	Chart       bool
}

func NewTestCmd() *cobra.Command {
	opts := &Options{}
	cfg := config.Get()

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Start a typing test",
		Long: `Start a typing test in one of three modes:

  timer  - Type as many words as you can before time runs out
  words  - Type a fixed number of words as fast as you can
  quote  - Type a famous quote

Examples:
  mtcli test                          # Default: 25 words
  mtcli test --mode timer --seconds 60  # 60 second timed test
  mtcli test --mode words --words 50    # Type 50 words
  mtcli test --mode quote --quote-random # Random quote`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(opts)
		},
	}

	// Mode flags
	cmd.Flags().StringVarP(&opts.Mode, "mode", "m", cfg.Mode, "test mode: timer, words, or quote")
	cmd.Flags().IntVarP(&opts.Seconds, "seconds", "s", cfg.Seconds, "duration in seconds (timer mode)")
	cmd.Flags().IntVarP(&opts.Words, "words", "w", cfg.Words, "number of words (words mode)")

	// Quote flags
	cmd.Flags().StringVar(&opts.QuoteID, "quote-id", "", "specific quote ID (quote mode)")
	cmd.Flags().BoolVar(&opts.QuoteRandom, "quote-random", true, "random quote (quote mode)")
	cmd.Flags().StringVar(&opts.QuotesFile, "quotes-file", cfg.QuotesFile, "custom quotes file")

	// Content flags
	cmd.Flags().StringVar(&opts.WordsFile, "words-file", cfg.WordsFile, "custom words file")

	// Behavior flags
	cmd.Flags().IntVar(&opts.Countdown, "countdown", cfg.Countdown, "countdown seconds before test starts")
	cmd.Flags().Int64Var(&opts.Seed, "seed", 0, "random seed for reproducible tests")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", cfg.NoColor, "disable color output")

	// Output flags
	cmd.Flags().IntVar(&opts.Wrap, "wrap", cfg.Wrap, "wrap width (0 for auto)")
	cmd.Flags().BoolVar(&opts.Chart, "chart", cfg.Chart, "show speed chart at end")

	return cmd
}

func runTest(opts *Options) error {
	// Create text generator
	gen, err := text.NewGenerator(text.GeneratorOptions{
		WordsFile:  opts.WordsFile,
		QuotesFile: opts.QuotesFile,
		Seed:       opts.Seed,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize text generator: %w", err)
	}

	// Generate target based on mode
	var target *test.Target
	switch opts.Mode {
	case "timer":
		target, err = gen.GenerateForTimer(opts.Seconds)
	case "words":
		target, err = gen.GenerateWords(opts.Words)
	case "quote":
		if opts.QuoteID != "" {
			target, err = gen.GetQuoteByID(opts.QuoteID)
		} else {
			target, err = gen.GetRandomQuote()
		}
	default:
		return fmt.Errorf("unknown mode: %s", opts.Mode)
	}

	if err != nil {
		return fmt.Errorf("failed to generate target text: %w", err)
	}

	// Create renderer
	renderer := ui.NewANSIRenderer(ui.RendererOptions{
		Width:   opts.Wrap,
		NoColor: opts.NoColor,
	})

	// Create input reader
	reader := input.NewRawReader()

	// Create session
	session := test.NewSession(test.SessionOptions{
		Target:       target,
		TimerSeconds: opts.Seconds,
	})

	// Initialize raw mode
	if err := reader.Init(); err != nil {
		return fmt.Errorf("failed to initialize input: %w", err)
	}
	defer reader.Cleanup()

	// Initialize renderer
	if err := renderer.Init(); err != nil {
		reader.Cleanup()
		return fmt.Errorf("failed to initialize renderer: %w", err)
	}
	defer renderer.Cleanup()

	// Countdown
	if opts.Countdown > 0 {
		for i := opts.Countdown; i > 0; i-- {
			renderer.RenderCountdown(i)
			time.Sleep(time.Second)
		}
	}

	// Initial render
	state := session.GetState()
	renderState := buildRenderState(session, state, opts)
	renderer.Render(renderState)

	// Channel for key events
	keyChan := make(chan input.KeyEvent)
	errChan := make(chan error)

	// Read input in goroutine
	go func() {
		for {
			key, err := reader.ReadKey()
			if err != nil {
				errChan <- err
				return
			}
			keyChan <- key
		}
	}()

	// Ticker for periodic updates (timer display, live WPM)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Main event loop
	for !session.IsFinished() {
		select {
		case key := <-keyChan:
			switch key.Type {
			case input.KeyCtrlC, input.KeyEscape:
				session.Abort()
			case input.KeyRune:
				session.HandleKey(test.KeyTypeRune, key.Rune)
			case input.KeyBackspace:
				session.HandleKey(test.KeyTypeBackspace, 0)
			}

			// Update display after keypress
			state = session.GetState()
			renderState = buildRenderState(session, state, opts)
			renderer.Render(renderState)

		case <-ticker.C:
			// Periodic update for timer mode and live WPM
			if !session.IsFinished() {
				// Collect sample for chart
				session.TakeSample()
				
				state = session.GetState()
				renderState = buildRenderState(session, state, opts)
				renderer.Render(renderState)
			}

		case err := <-errChan:
			return fmt.Errorf("input error: %w", err)
		}
	}

	// If aborted, exit without summary
	if session.IsAborted() {
		return nil
	}

	// Get results
	result := session.GetResult()

	// Generate chart
	var chartStr string
	if opts.Chart && len(result.Samples) > 1 {
		// Convert samples to chart data points
		wpmPoints := make([]charts.DataPoint, len(result.Samples))
		rawPoints := make([]charts.DataPoint, len(result.Samples))
		for i, s := range result.Samples {
			wpmPoints[i] = charts.DataPoint{TimeMs: s.TimeMs, Value: s.WPM}
			rawPoints[i] = charts.DataPoint{TimeMs: s.TimeMs, Value: s.RawWPM}
		}

		chartOpts := charts.DefaultOptions()
		chartOpts.Width = renderer.GetWidth() - 4
		if chartOpts.Width > 70 {
			chartOpts.Width = 70
		}
		chartStr = charts.RenderDualChart(wpmPoints, rawPoints, chartOpts)
	}

	// Show summary
	renderer.RenderSummary(result, chartStr)

	// Save to storage
	if err := saveSession(result); err != nil {
		fmt.Printf("Warning: failed to save session: %v\n", err)
	}

	return nil
}

func buildRenderState(session *test.Session, state *test.SessionState, opts *Options) *ui.RenderState {
	return &ui.RenderState{
		Target:     state.TargetRunes,
		Typed:      state.TypedRunes,
		CharStates: state.CharStates,
		Mode:       state.Target.Mode,
		Elapsed:    session.GetElapsed().Seconds(),
		LiveWPM:    session.GetLiveWPM(),
		TimeLimit:  opts.Seconds,
		Finished:   state.Finished,
	}
}

func saveSession(result *test.SessionResult) error {
	store, err := sqlite.Open()
	if err != nil {
		return err
	}
	defer store.Close()

	// Convert to storage types
	session := &sqlite.Session{
		StartedAt:    result.StartedAt,
		Mode:         string(result.Mode),
		Seconds:      result.Metadata.Seconds,
		Words:        result.Metadata.WordCount,
		QuoteID:      result.Metadata.QuoteID,
		TargetLen:    result.TargetLen,
		DurationMs:   result.Duration.Milliseconds(),
		CorrectChars: result.CorrectChars,
		TotalTyped:   result.TotalTyped,
		Accuracy:     result.Accuracy,
		WPM:          result.WPM,
		RawWPM:       result.RawWPM,
	}

	samples := make([]sqlite.SessionSample, len(result.Samples))
	for i, s := range result.Samples {
		samples[i] = sqlite.SessionSample{
			TimeMs: s.TimeMs,
			WPM:    s.WPM,
			RawWPM: s.RawWPM,
		}
	}

	_, err = store.SaveSession(session, samples)
	return err
}
