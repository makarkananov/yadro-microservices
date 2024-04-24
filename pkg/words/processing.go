package words

import (
	"errors"
	"fmt"
	"github.com/bbalet/stopwords"
	"github.com/kljensen/snowball"
	"strings"
	"unicode"
)

// TextProcessor is a structure for text processing.
// It contains the language code and the file path for stop words.
type TextProcessor struct {
	lang string // Language code for processing
}

// NewTextProcessor creates a new instance of TextProcessor.
// It takes a language code and a file path for stop words as parameters.
func NewTextProcessor(lang string, stopWordsFile string) *TextProcessor {
	if stopWordsFile != "" {
		stopwords.LoadStopWordsFromFile(stopWordsFile, lang, " ")
	}

	return &TextProcessor{lang: lang}
}

// FullProcess performs the full cycle of text processing.
// It tokenizes the text, removes stop words and normalizes the tokens.
func (tp *TextProcessor) FullProcess(text string) ([]string, error) {
	tokens := tp.Tokenize(text)

	tokens, err := tp.RemoveStopWords(tokens)
	if err != nil {
		return nil, fmt.Errorf("error while removing stop words: %w", err)
	}

	tokens, err = tp.Normalize(tokens)
	if err != nil {
		return nil, fmt.Errorf("error while normalizing: %w", err)
	}

	return tokens, nil
}

// Tokenize breaks down the text into tokens.
// It splits the text by punctuation and spaces, but treats apostrophes as part of the word.
// If an apostrophe is found, it removes apostrophe with the rest of the word after the apostrophe.
func (tp *TextProcessor) Tokenize(text string) []string {
	f := func(c rune) bool {
		return (unicode.IsPunct(c) && c != '\'') || unicode.IsSpace(c)
	}

	tokens := strings.FieldsFunc(text, f)

	for i, token := range tokens {
		if index := strings.IndexRune(token, '\''); index != -1 {
			tokens[i] = token[:index]
		}
	}

	return tokens
}

// Normalize normalizes the token slice by stemming each token.
// It uses the snowball library for stemming.
func (tp *TextProcessor) Normalize(tokens []string) ([]string, error) {
	fullLangName, ok := languageCodesMap[tp.lang]
	if !ok {
		return nil, errors.New("unsupported language code")
	}

	normalizedWords := make([]string, 0, len(tokens))
	for _, token := range tokens {
		stemmed, err := snowball.Stem(token, fullLangName, true)
		if err != nil {
			return nil, fmt.Errorf("error during stemming: %w", err)
		}
		normalizedWords = append(normalizedWords, stemmed)
	}

	return normalizedWords, nil
}

// RemoveStopWords removes stop words from the token slice.
// It uses the stopwords library to clean the text.
func (tp *TextProcessor) RemoveStopWords(tokens []string) ([]string, error) {
	if _, ok := languageCodesMap[tp.lang]; !ok {
		return nil, errors.New("unsupported language code")
	}

	cleanedText := stopwords.CleanString(strings.Join(tokens, " "), tp.lang, false)
	return strings.Fields(cleanedText), nil
}

// RemoveDuplicates removes duplicates from the token slice.
// It uses a map to track encountered tokens and only keeps the unique ones.
func (tp *TextProcessor) RemoveDuplicates(tokens []string) []string {
	encountered := map[string]bool{}
	var uniqueTokens []string
	for _, token := range tokens {
		if !encountered[token] {
			encountered[token] = true
			uniqueTokens = append(uniqueTokens, token)
		}
	}

	return uniqueTokens
}

// languageCodesMap maps language codes to their full names.
// It is used to check if a language code is supported and to get the full language name for stemming.
var languageCodesMap = map[string]string{
	"en": "english",
	"ru": "russian",
	"fr": "french",
	"es": "spanish",
	"sv": "swedish",
}
