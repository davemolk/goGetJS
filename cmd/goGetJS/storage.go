package main

import "sync"

// SearchMap is a Mutex-protected struct that stores search results
// in the following format: script url: found search term(s)
type SearchMap struct {
	mu       sync.Mutex
	Searches map[string][]string
}

// NewSearchMap creates a new SearchMap and returns a pointer to it.
func NewSearchMap() *SearchMap {
	return &SearchMap{
		Searches: make(map[string][]string),
	}
}

// Store receives a script url and the term found on that page
// and records it in the SearchMap.
func (s *SearchMap) Store(url, term string) {
	s.mu.Lock()
	s.Searches[url] = append(s.Searches[url], term)
	s.mu.Unlock()
}
