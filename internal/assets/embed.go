package assets

import (
	_ "embed"
)

//go:embed words.txt
var WordsData string

//go:embed quotes.json
var QuotesData string

