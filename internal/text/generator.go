package text

import (
	"fmt"

	"github.com/mmdbasi/mtcli/internal/test"
)

// DefaultGenerator implements text generation for all modes
type DefaultGenerator struct {
	wordList  *WordList
	quoteList *QuoteList
}

// GeneratorOptions holds configuration for the generator
type GeneratorOptions struct {
	WordsFile  string
	QuotesFile string
	Seed       int64
}

// NewGenerator creates a new text generator
func NewGenerator(opts GeneratorOptions) (*DefaultGenerator, error) {
	wordList, err := NewWordList(opts.WordsFile, opts.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to load words: %w", err)
	}

	quoteList, err := NewQuoteList(opts.QuotesFile, opts.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to load quotes: %w", err)
	}

	return &DefaultGenerator{
		wordList:  wordList,
		quoteList: quoteList,
	}, nil
}

// GenerateWords generates a target with the specified number of words
func (g *DefaultGenerator) GenerateWords(count int) (*test.Target, error) {
	if count <= 0 {
		return nil, fmt.Errorf("word count must be positive")
	}

	text := g.wordList.GenerateText(count)

	return &test.Target{
		Text: text,
		Mode: test.ModeWords,
		Metadata: test.TargetMetadata{
			WordCount: count,
		},
	}, nil
}

// GenerateForTimer generates enough words for a timed test
// Assumes average typing speed of ~200 WPM (very fast) to ensure enough words
func (g *DefaultGenerator) GenerateForTimer(seconds int) (*test.Target, error) {
	if seconds <= 0 {
		return nil, fmt.Errorf("seconds must be positive")
	}

	// Generate enough words for even fast typists (200 WPM = ~3.3 words/sec)
	// Add 50% buffer to be safe
	wordCount := int(float64(seconds) * 4)
	if wordCount < 50 {
		wordCount = 50
	}

	text := g.wordList.GenerateText(wordCount)

	return &test.Target{
		Text: text,
		Mode: test.ModeTimer,
		Metadata: test.TargetMetadata{
			Seconds:   seconds,
			WordCount: wordCount,
		},
	}, nil
}

// GetRandomQuote returns a random quote as a target
func (g *DefaultGenerator) GetRandomQuote() (*test.Target, error) {
	quote := g.quoteList.GetRandomQuote()
	if quote == nil {
		return nil, fmt.Errorf("no quotes available")
	}

	return &test.Target{
		Text: quote.Text,
		Mode: test.ModeQuote,
		Metadata: test.TargetMetadata{
			QuoteID: quote.ID,
			Source:  quote.Source,
		},
	}, nil
}

// GetQuoteByID returns a specific quote as a target
func (g *DefaultGenerator) GetQuoteByID(id string) (*test.Target, error) {
	quote, err := g.quoteList.GetQuoteByID(id)
	if err != nil {
		return nil, err
	}

	return &test.Target{
		Text: quote.Text,
		Mode: test.ModeQuote,
		Metadata: test.TargetMetadata{
			QuoteID: quote.ID,
			Source:  quote.Source,
		},
	}, nil
}

