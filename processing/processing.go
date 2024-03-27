package processing

import (
	"errors"
	"fmt"
	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"
	"strings"
)

// TextProcessor is a structure for text processing
type TextProcessor struct {
	tokens []string // Tokens extracted from the text
	lang   string   // Language code for processing
}

// NewTextProcessor creates a new instance of TextProcessor
func NewTextProcessor(text string, lang string) *TextProcessor {
	tp := &TextProcessor{lang: lang}
	tp.tokenize(text)
	return tp
}

// tokenize breaks down the text into tokens
func (tp *TextProcessor) tokenize(text string) {
	words := strings.Fields(text)
	for i, word := range words {
		if index := strings.Index(word, "'"); index != -1 {
			words[i] = word[:index]
		}
	}
	tp.tokens = words
}

// GetText returns the final processed text
func (tp *TextProcessor) GetText() string {
	return strings.Join(tp.tokens, " ")
}

// Normalize normalizes the text by stemming each token
func (tp *TextProcessor) Normalize() error {
	fullLangName, ok := languageCodesMap[tp.lang]
	if !ok {
		return errors.New("unsupported language code")
	}

	var normalizedWords []string
	for _, word := range tp.tokens {
		stemmed, err := snowball.Stem(word, fullLangName, true)
		if err != nil {
			return fmt.Errorf("error during normalization: %w", err)
		}
		normalizedWords = append(normalizedWords, stemmed)
	}

	tp.tokens = normalizedWords
	return nil
}

// RemoveStopWords removes stop words from the text
func (tp *TextProcessor) RemoveStopWords(stopWordsFile string) error {
	if _, ok := languageCodesMap[tp.lang]; !ok {
		return errors.New("unsupported language code")
	}

	if stopWordsFile != "" {
		stopwords.LoadStopWordsFromFile(stopWordsFile, tp.lang, " ")
	}

	tp.tokens = strings.Fields(stopwords.CleanString(tp.GetText(), tp.lang, false))
	return nil
}

// RemoveDuplicates removes duplicates from the token slice
func (tp *TextProcessor) RemoveDuplicates() {
	encountered := map[string]bool{}
	var uniqueTokens []string
	for _, token := range tp.tokens {
		if !encountered[token] {
			encountered[token] = true
			uniqueTokens = append(uniqueTokens, token)
		}
	}
	tp.tokens = uniqueTokens
}

// languageCodesMap maps language codes to their full names
var languageCodesMap = map[string]string{
	"en": "english",
	"ru": "russian",
	"fr": "french",
	"es": "spanish",
	"sv": "swedish",
}
