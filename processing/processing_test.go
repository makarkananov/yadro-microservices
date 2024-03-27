package processing

import (
	"testing"
)

func TestTextProcessor_GetText(t *testing.T) {
	text := "hello world"
	lang := "en"
	tp := NewTextProcessor(text, lang)

	got := tp.GetText()
	want := "hello world"

	if got != want {
		t.Errorf("GetText() = %q, want %q", got, want)
	}
}

func TestTextProcessor_Normalize(t *testing.T) {
	text := "running fast"
	lang := "en"
	tp := NewTextProcessor(text, lang)

	err := tp.Normalize()
	if err != nil {
		t.Fatalf("Normalize() error: %v", err)
	}

	got := tp.GetText()
	want := "run fast"

	if got != want {
		t.Errorf("Normalize() = %q, want %q", got, want)
	}
}

func TestTextProcessor_RemoveStopWords(t *testing.T) {
	text := "the quick brown fox"
	lang := "en"
	stopWordsFile := ""
	tp := NewTextProcessor(text, lang)

	err := tp.RemoveStopWords(stopWordsFile)
	if err != nil {
		t.Fatalf("RemoveStopWords() error: %v", err)
	}

	got := tp.GetText()
	want := "quick brown fox"

	if got != want {
		t.Errorf("RemoveStopWords() = %q, want %q", got, want)
	}
}

func TestTextProcessor_RemoveDuplicates(t *testing.T) {
	text := "hello world hello world"
	lang := "en"
	tp := NewTextProcessor(text, lang)

	tp.RemoveDuplicates()

	got := tp.GetText()
	want := "hello world"

	if got != want {
		t.Errorf("RemoveDuplicates() = %q, want %q", got, want)
	}
}

func TestTextProcessor_NormalizeUnsupportedLanguage(t *testing.T) {
	text := "hello world"
	lang := "de" // Unsupported language
	tp := NewTextProcessor(text, lang)

	err := tp.Normalize()

	if err == nil {
		t.Errorf("Normalize() failed: expected error for unsupported language, got nil")
	}
}

func TestTextProcessor_RemoveStopWordsUnsupportedLanguage(t *testing.T) {
	text := "the quick brown fox"
	lang := "de" // Unsupported language
	stopWordsFile := ""
	tp := NewTextProcessor(text, lang)

	err := tp.RemoveStopWords(stopWordsFile)

	if err == nil {
		t.Errorf("RemoveStopWords() failed: expected error for unsupported language, got nil")
	}
}
