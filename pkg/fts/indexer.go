package fts

// Index is a struct that represents the score for document with specific ID.
type Index struct {
	ID    int `json:"id"`
	Score int `json:"counter"`
}

// Indexer is an interface that defines the behavior of an indexing engine.
type Indexer interface {
	Add(doc *Document)
	Get(token string) []*Index
}

// InvertedIndexer is an implementation of the Indexer interface that uses an inverted index.
type InvertedIndexer struct {
	Indexes          map[string][]*Index `json:"indexes"`
	IndexedDocuments map[int]bool        `json:"indexed-documents"`
}

// NewInvertedIndexer creates a new InvertedIndexer.
func NewInvertedIndexer() *InvertedIndexer {
	return &InvertedIndexer{
		Indexes:          make(map[string][]*Index),
		IndexedDocuments: make(map[int]bool),
	}
}

// Add adds a document to the inverted index.
func (i *InvertedIndexer) Add(doc *Document) {
	if i.IndexedDocuments[doc.ID] { // If the document is already indexed, skip it
		return
	}

	for _, token := range doc.Tokens {
		if _, ok := i.Indexes[token]; !ok {
			i.Indexes[token] = []*Index{
				{
					ID:    doc.ID,
					Score: 1,
				},
			}

			continue
		}

		found := false
		for _, index := range i.Indexes[token] {
			if index.ID == doc.ID {
				found = true
				index.Score++
			}
		}

		if !found {
			i.Indexes[token] = append(i.Indexes[token], &Index{
				ID:    doc.ID,
				Score: 1,
			})
		}
	}

	i.IndexedDocuments[doc.ID] = true
}

// Get returns the indexes for a given token.
func (i *InvertedIndexer) Get(token string) []*Index {
	return i.Indexes[token]
}
