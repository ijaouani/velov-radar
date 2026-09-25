package stationcache

import (
	"fmt"
	"sync"

	"velov-radar/velov"
)

type Cache struct {
	client  *velov.Client
	mu      sync.RWMutex
	station map[int]*velov.Station
}

func NewCache(client *velov.Client) *Cache {
	return &Cache{
		client:  client,
		station: make(map[int]*velov.Station),
	}
}

func (c *Cache) Refresh() error {
	stations, err := c.client.FetchStations()
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range stations {
		c.station[stations[i].Number] = &stations[i]
	}
	return nil
}

func (c *Cache) GetStation(stationID int) (*velov.Station, error) {
	c.mu.RLock()
	targetStation, ok := c.station[stationID]
	c.mu.RUnlock()

	if ok {
		return targetStation, nil
	}

	// Cache miss, force refresh
	if err := c.Refresh(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	targetStation, ok = c.station[stationID]
	c.mu.RUnlock()

	if ok {
		return targetStation, nil
	}

	return nil, fmt.Errorf("station not found")
}
