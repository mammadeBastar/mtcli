package text

import "github.com/mmdbasi/mtcli/internal/test"

// Generator defines the interface for text generation
type Generator interface {
	// GenerateWords generates a random word sequence
	GenerateWords(count int) (*test.Target, error)

	// GenerateForTimer generates enough words for a timed test
	GenerateForTimer(seconds int) (*test.Target, error)

	// GetRandomQuote returns a random quote
	GetRandomQuote() (*test.Target, error)

	// GetQuoteByID returns a specific quote
	GetQuoteByID(id string) (*test.Target, error)
}

// Quote represents a quote with metadata
type Quote struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

