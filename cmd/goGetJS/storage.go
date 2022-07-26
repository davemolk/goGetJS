package main

import "sync"

type SearchMap struct {
	mu       sync.RWMutex
	Searches map[string][]string // url: term(s)
}

func NewSearchMap() *SearchMap {
	return &SearchMap{
		Searches: make(map[string][]string),
	}
}

func (rw *SearchMap) Load(key string) ([]string, bool) {
	rw.mu.RLock()
	result, ok := rw.Searches[key]
	rw.mu.RUnlock()
	return result, ok
}

func (rw *SearchMap) Store(url string, term string) {
	rw.mu.Lock()
	rw.Searches[url] = append(rw.Searches[url], term)
	rw.mu.Unlock()
}
