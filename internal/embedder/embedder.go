// Package embedder provides text embedding for semantic search.
// Uses TF-IDF vectorization which requires no external ML dependencies
// while still providing effective semantic code search.
package embedder

import (
	"context"
	"math"
	"regexp"
	"strings"
	"sync"
)

// Embedder provides TF-IDF based text embedding.
// This approach requires no external ML models and works entirely in-memory.
type Embedder struct {
	vocab     map[string]int // word -> index
	docFreq   map[string]int // word -> number of docs containing it
	docCount  int            // total number of documents indexed
	vocabSize int            // number of unique terms
	mu        sync.RWMutex
}

// New creates a new TF-IDF embedder.
func New() *Embedder {
	return &Embedder{
		vocab:   make(map[string]int),
		docFreq: make(map[string]int),
	}
}

// tokenize splits text into normalized tokens.
func tokenize(text string) []string {
	// Lowercase and split on non-alphanumeric
	text = strings.ToLower(text)
	re := regexp.MustCompile(`[a-z0-9_]+`)
	return re.FindAllString(text, -1)
}

// Stop words to exclude from indexing
var stopWords = map[string]bool{
	"the": true, "and": true, "or": true, "but": true, "in": true,
	"on": true, "at": true, "to": true, "for": true, "of": true,
	"is": true, "it": true, "that": true, "this": true, "with": true,
	"from": true, "by": true, "an": true, "be": true, "are": true,
	"was": true, "were": true, "not": true, "if": true, "then": true,
	"else": true, "do": true, "does": true, "has": true, "have": true,
	"will": true, "would": true, "could": true, "should": true,
}

// Embed returns a TF-IDF vector for the given text.
// The vector is normalized to unit length for cosine similarity.
func (e *Embedder) Embed(ctx context.Context, text string) []float32 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tokens := tokenize(text)
	if len(tokens) == 0 {
		return nil
	}

	// Term frequency
	tf := make(map[string]int)
	for _, t := range tokens {
		if stopWords[t] {
			continue
		}
		tf[t]++
	}

	// TF-IDF vector
	vocabSize := len(e.vocab)
	if vocabSize == 0 {
		return nil
	}
	vec := make([]float32, vocabSize)
	var norm float64

	for term, count := range tf {
		idx, ok := e.vocab[term]
		if !ok {
			continue
		}
		// TF-IDF: term_freq * log(total_docs / doc_freq)
		df := float64(e.docFreq[term])
		if df == 0 {
			df = 1
		}
		idf := math.Log(float64(e.docCount) / df)
		tfidf := float64(count) * idf
		vec[idx] = float32(tfidf)
		norm += tfidf * tfidf
	}

	// L2 normalize
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= float32(norm)
		}
	}

	return vec
}

// IndexDocument adds a document to the vocabulary and updates document frequencies.
// Must be called before Embed to build the vocabulary.
func (e *Embedder) IndexDocument(text string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	tokens := tokenize(text)
	seen := make(map[string]bool)
	for _, t := range tokens {
		if stopWords[t] {
			continue
		}
		if !seen[t] {
			seen[t] = true
			if _, ok := e.vocab[t]; !ok {
				e.vocab[t] = len(e.vocab)
			}
			e.docFreq[t]++
		}
	}
	e.docCount++
	e.vocabSize = len(e.vocab)
}

// IndexDocuments batch-indexes multiple documents.
func (e *Embedder) IndexDocuments(texts []string) {
	for _, text := range texts {
		e.IndexDocument(text)
	}
}

// CosineSimilarity computes the cosine similarity between two pre-computed vectors.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

// VocabSize returns the number of unique terms in the vocabulary.
func (e *Embedder) VocabSize() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vocabSize
}

// DocCount returns the number of indexed documents.
func (e *Embedder) DocCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.docCount
}
