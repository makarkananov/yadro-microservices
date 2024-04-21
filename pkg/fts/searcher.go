package fts

import (
	"sort"
)

// Searcher is an interface that defines the behavior of a search engine.
type Searcher interface {
	Search(queryTokens []string, modifiers ...SearchModifier) SearchResults
}

// SearchModifier is a function that modifies the search results.
type SearchModifier func([]string, *SearchResults)

type SearchResult struct {
	ID             int // ID of the document
	NumberOfTokens int // Number of tokens matched with query
	Score          int // Total number of occurrences of the tokens
}

// SearchResults implements the sort.Interface.
type SearchResults []*SearchResult

func (s SearchResults) Len() int { return len(s) }

func (s SearchResults) Less(i, j int) bool {
	if s[i].NumberOfTokens == s[j].NumberOfTokens {
		if s[i].Score == s[j].Score {
			return s[i].ID < s[j].ID
		} // Third priority is the ID of the document

		return s[i].Score > s[j].Score // Second priority is the total number of occurrences of different tokens
	}

	return s[i].NumberOfTokens > s[j].NumberOfTokens // First priority is the number of distinct matched tokens
}

func (s SearchResults) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s SearchResults) FindByID(id int) *SearchResult {
	for _, sr := range s {
		if sr.ID == id {
			return sr
		}
	}

	return nil
}

// TokenResult is a struct that represents the number of occurrences of a token in a document with specific ID.
type TokenResult struct {
	ID    int
	Score int
}

// FullTextSearcher is an implementation of Searcher.
type FullTextSearcher struct{}

// Search searches the query tokens by applying the modifiers to the search results.
func (s *FullTextSearcher) Search(queryTokens []string, modifiers ...SearchModifier) SearchResults {
	var searchResults SearchResults
	for _, modifier := range modifiers {
		modifier(queryTokens, &searchResults)
	}

	return searchResults
}

// ReturnMostRelevant returns the most relevant n search results.
func ReturnMostRelevant(n int) SearchModifier {
	return func(_ []string, results *SearchResults) {
		sort.Sort(results)
		if len(*results) > n {
			*results = (*results)[:n]
		}
	}
}

// ThroughIndexes is a search modifier that searches using the indexer.
func ThroughIndexes(indexer Indexer) SearchModifier {
	return func(queryTokens []string, results *SearchResults) {
		for _, token := range queryTokens {
			tokenResults := indexer.Get(token)
			for _, tr := range tokenResults {
				r := results.FindByID(tr.ID)
				if r == nil {
					r = &SearchResult{
						NumberOfTokens: 1,
						Score:          tr.Score,
						ID:             tr.ID,
					}
					*results = append(*results, r)
					continue
				}

				r.NumberOfTokens++
				r.Score += tr.Score
			}
		}
	}
}

// ThroughDocs is a search modifier that searches using the documents.
func ThroughDocs(docs []*Document) SearchModifier {
	return func(queryTokens []string, results *SearchResults) {
		for _, token := range queryTokens {
			tokenResults := searchToken(docs, token)
			for _, tr := range tokenResults {
				r := results.FindByID(tr.ID)
				if r == nil {
					r = &SearchResult{
						NumberOfTokens: 1,
						Score:          tr.Score,
						ID:             tr.ID,
					}
					*results = append(*results, r)
					continue
				}

				r.NumberOfTokens++
				r.Score += tr.Score
			}
		}
	}
}

func searchToken(docs []*Document, token string) []*TokenResult {
	var results []*TokenResult

	for _, doc := range docs {
		result := &TokenResult{
			ID:    doc.ID,
			Score: 0,
		}

		found := false
		for _, t := range doc.Tokens {
			if t == token {
				result.Score++
				found = true
			}
		}

		if found {
			results = append(results, result)
		}
	}

	return results
}
