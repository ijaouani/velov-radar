package store

import (
	"encoding/json"
	"os"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	file string
	Data map[int64][]int // chatID -> []stationID
}

func NewStore(file string) (*Store, error) {
	s := &Store{
		file: file,
		Data: make(map[int64][]int),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, start fresh
		}
		return err
	}

	if err := json.Unmarshal(b, &s.Data); err != nil {
		return err
	}
	return nil
}

func (s *Store) save() {
	b, _ := json.MarshalIndent(s.Data, "", "  ")
	os.WriteFile(s.file, b, 0644)
}

func (s *Store) Add(chatID int64, stationID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check for duplicates
	for _, id := range s.Data[chatID] {
		if id == stationID {
			return // Already exists
		}
	}
	s.Data[chatID] = append(s.Data[chatID], stationID)
	s.save()
}

func (s *Store) Remove(chatID int64, stationID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stations := s.Data[chatID]
	for i, id := range stations {
		if id == stationID {
			s.Data[chatID] = append(stations[:i], stations[i+1:]...)
			s.save()
			break
		}
	}
}

func (s *Store) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Data, chatID)
	s.save()
}

func (s *Store) GetAll() map[int64][]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[int64][]int)
	for k, v := range s.Data {
		cp[k] = append([]int(nil), v...) // Deep copy
	}
	return cp
}
