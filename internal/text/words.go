package text

import (
	"bufio"
	"math/rand"
	"os"
	"strings"

	"github.com/mmdbasi/mtcli/internal/assets"
)

// WordList holds a list of words for generating typing tests
type WordList struct {
	words []string
	rng   *rand.Rand
}

// NewWordList creates a new word list from the embedded words or a custom file
func NewWordList(customFile string, seed int64) (*WordList, error) {
	var words []string
	var err error

	if customFile != "" {
		words, err = loadWordsFromFile(customFile)
	} else {
		words, err = loadEmbeddedWords()
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

	return &WordList{
		words: words,
		rng:   rng,
	}, nil
}

// loadEmbeddedWords loads words from the embedded words.txt
func loadEmbeddedWords() ([]string, error) {
	var words []string
	scanner := bufio.NewScanner(strings.NewReader(assets.WordsData))
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}
	return words, scanner.Err()
}

// loadWordsFromFile loads words from a custom file
func loadWordsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}
	return words, scanner.Err()
}

// GetRandomWords returns n random words
func (wl *WordList) GetRandomWords(n int) []string {
	if n <= 0 {
		return nil
	}

	result := make([]string, n)
	for i := 0; i < n; i++ {
		idx := wl.rng.Intn(len(wl.words))
		result[i] = wl.words[idx]
	}
	return result
}

// GenerateText generates a text string of n random words
func (wl *WordList) GenerateText(n int) string {
	words := wl.GetRandomWords(n)
	return strings.Join(words, " ")
}

// Count returns the number of words in the list
func (wl *WordList) Count() int {
	return len(wl.words)
}

