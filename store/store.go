package store

import (
	"encoding/json"
	"os"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	file string
	Data map[int64]int // chatID -> stationID
}

func NewStore(file string) (*Store, error) {
	s := &Store{
		file: file,
		Data: make(map[int64]int),
	}
	s.load()
	return s, nil
}

func (s *Store) load() {
	b, err := os.ReadFile(s.file)
	if err == nil {
		json.Unmarshal(b, &s.Data)
	}
}

func (s *Store) save() {
	b, _ := json.MarshalIndent(s.Data, "", "  ")
	os.WriteFile(s.file, b, 0644)
}

func (s *Store) Set(chatID int64, stationID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Data[chatID] = stationID
	s.save()
}

func (s *Store) Delete(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Data, chatID)
	s.save()
}

func (s *Store) GetAll() map[int64]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[int64]int)
	for k, v := range s.Data {
		cp[k] = v
	}
	return cp
}
