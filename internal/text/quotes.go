package text

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/mmdbasi/mtcli/internal/assets"
)

// QuoteList holds a list of quotes
type QuoteList struct {
	quotes []Quote
	rng    *rand.Rand
}

// NewQuoteList creates a new quote list from embedded quotes or a custom file
func NewQuoteList(customFile string, seed int64) (*QuoteList, error) {
	var quotes []Quote
	var err error

	if customFile != "" {
		quotes, err = loadQuotesFromFile(customFile)
	} else {
		quotes, err = loadEmbeddedQuotes()
	}

	if err != nil {
		return nil, err
	}

	// Use provided seed or current time
	var rng *rand.Rand
	if seed != 0 {
		rng = rand.New(rand.NewSource(seed))
	} else {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	return &QuoteList{
		quotes: quotes,
		rng:    rng,
	}, nil
}

// loadEmbeddedQuotes loads quotes from embedded data
func loadEmbeddedQuotes() ([]Quote, error) {
	var quotes []Quote
	err := json.Unmarshal([]byte(assets.QuotesData), &quotes)
	return quotes, err
}

// loadQuotesFromFile loads quotes from a custom JSON file
func loadQuotesFromFile(path string) ([]Quote, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var quotes []Quote
	err = json.Unmarshal(data, &quotes)
	return quotes, err
}

// GetRandomQuote returns a random quote
func (ql *QuoteList) GetRandomQuote() *Quote {
	if len(ql.quotes) == 0 {
		return nil
	}
	idx := ql.rng.Intn(len(ql.quotes))
	return &ql.quotes[idx]
}

// GetQuoteByID returns a quote by its ID
func (ql *QuoteList) GetQuoteByID(id string) (*Quote, error) {
	for _, q := range ql.quotes {
		if q.ID == id {
			return &q, nil
		}
	}
	return nil, fmt.Errorf("quote with ID %q not found", id)
}

// Count returns the number of quotes in the list
func (ql *QuoteList) Count() int {
	return len(ql.quotes)
}

// ListIDs returns all available quote IDs
func (ql *QuoteList) ListIDs() []string {
	ids := make([]string, len(ql.quotes))
	for i, q := range ql.quotes {
		ids[i] = q.ID
	}
	return ids
}

