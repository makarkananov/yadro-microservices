package fts

// Document represents a document that can be indexed or searched for.
type Document struct {
	ID     int
	Tokens []string
}
