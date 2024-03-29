package processing

import (
	"reflect"
	"testing"
)

func TestTextProcessor_Tokenize(t *testing.T) {
	tp := NewTextProcessor("en", "")

	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "Tokenize with punctuation and spaces",
			text: "Hello, world! How are you?",
			want: []string{"Hello", "world", "How", "are", "you"},
		},
		{
			name: "Tokenize with hyphen",
			text: "end-to-end",
			want: []string{"end", "to", "end"},
		},
		{
			name: "Tokenize with apostrophe",
			text: "I'm fine",
			want: []string{"I", "fine"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tp.Tokenize(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tokenize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessor_Normalize(t *testing.T) {
	tp := NewTextProcessor("en", "")

	tests := []struct {
		name    string
		tokens  []string
		want    []string
		wantErr bool
	}{
		{
			name:    "Normalize English words 1",
			tokens:  []string{"running", "fast"},
			want:    []string{"run", "fast"},
			wantErr: false,
		},
		{
			name:    "Normalize English words 2",
			tokens:  []string{"following", "you"},
			want:    []string{"follow", "you"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tp.Normalize(tt.tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessor_RemoveStopWords(t *testing.T) {
	tp := NewTextProcessor("en", "")

	tests := []struct {
		name    string
		text    []string
		want    []string
		wantErr bool
	}{
		{
			name:    "Remove stop words from English text",
			text:    []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"},
			want:    []string{"quick", "brown", "fox", "jumps", "lazy", "dog"},
			wantErr: false,
		},
		{
			name:    "Remove stop words from empty text",
			text:    []string{},
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "Remove stop words from text with only stop words",
			text:    []string{"the", "the", "the"},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tp.RemoveStopWords(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveStopWords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStopWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessor_RemoveDuplicates(t *testing.T) {
	tp := NewTextProcessor("en", "")

	tests := []struct {
		name   string
		tokens []string
		want   []string
	}{
		{
			name:   "Remove duplicates from token slice",
			tokens: []string{"hello", "world", "hello", "world"},
			want:   []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tp.RemoveDuplicates(tt.tokens); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessor_FullProcess(t *testing.T) {
	tp := NewTextProcessor("en", "")

	tests := []struct {
		name    string
		text    string
		want    []string
		wantErr bool
	}{
		{
			name:    "Full process with English text",
			text:    "The quick brown fox jumps over the lazy dog",
			want:    []string{"quick", "brown", "fox", "jump", "lazi", "dog"},
			wantErr: false,
		},
		{
			name:    "Full process with empty text",
			text:    "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Full process with text containing only stop words",
			text:    "the the the",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Full process with text containing apostrophes",
			text:    "I'm fine, it's a beautiful day",
			want:    []string{"i", "fine", "beauti", "day"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tp.FullProcess(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("FullProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FullProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessor_FullProcess_Russian(t *testing.T) {
	tp := NewTextProcessor("ru", "")

	tests := []struct {
		name    string
		text    string
		want    []string
		wantErr bool
	}{
		{
			name:    "Full process with Russian text",
			text:    "Быстрый коричневый лис прыгает через ленивую собаку",
			want:    []string{"быстр", "коричнев", "лис", "прыга", "ленив", "собак"},
			wantErr: false,
		},
		{
			name:    "Full process with Russian text containing only stop words",
			text:    "и и и",
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tp.FullProcess(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("FullProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FullProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}
