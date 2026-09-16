package monitor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"velov-radar/notifier"
	"velov-radar/store"
	"velov-radar/velov"
)

type StationState struct {
	HasElec bool
	HasMeca bool
}

type Monitor struct {
	client       *velov.Client
	notifier     *notifier.TelegramNotifier
	store        *store.Store
	pollInterval time.Duration
	states       map[int]*StationState

	mu           sync.RWMutex
	stationCache map[int]*velov.Station
}

func NewMonitor(client *velov.Client, notif *notifier.TelegramNotifier, store *store.Store, interval time.Duration) *Monitor {
	return &Monitor{
		client:       client,
		notifier:     notif,
		store:        store,
		pollInterval: interval,
		states:       make(map[int]*StationState),
		stationCache: make(map[int]*velov.Station),
	}
}

func (m *Monitor) Start() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	// Initial fetch to set up state without notifying
	m.check()

	for range ticker.C {
		m.check()
	}
}

func (m *Monitor) refreshCache() error {
	stations, err := m.client.FetchStations()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range stations {
		m.stationCache[stations[i].Number] = &stations[i]
	}
	return nil
}

func (m *Monitor) GetStation(stationID int) (*velov.Station, error) {
	m.mu.RLock()
	targetStation, ok := m.stationCache[stationID]
	m.mu.RUnlock()

	if ok {
		return targetStation, nil
	}

	// Cache miss, force refresh
	if err := m.refreshCache(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	targetStation, ok = m.stationCache[stationID]
	m.mu.RUnlock()

	if ok {
		return targetStation, nil
	}

	return nil, fmt.Errorf("station not found")
}

func (m *Monitor) check() {
	if err := m.refreshCache(); err != nil {
		log.Printf("Error fetching stations: %v", err)
		return
	}

	subs := m.store.GetAll()
	if len(subs) == 0 {
		// Clear states so we don't keep stale data if everyone unsubscribes
		m.states = make(map[int]*StationState)
		return // No users subscribed yet
	}

	// Figure out which stations are being monitored
	monitoredStations := make(map[int][]int64) // stationID -> []chatID
	for chatID, stationIDs := range subs {
		for _, stationID := range stationIDs {
			monitoredStations[stationID] = append(monitoredStations[stationID], chatID)
		}
	}

	// Clean up unmonitored stations from state
	for stationID := range m.states {
		if _, ok := monitoredStations[stationID]; !ok {
			delete(m.states, stationID)
		}
	}

	for stationID, chatIDs := range monitoredStations {
		m.mu.RLock()
		targetStation, ok := m.stationCache[stationID]
		m.mu.RUnlock()

		if !ok {
			log.Printf("Station %d not found in API response", stationID)
			continue
		}

		elec := targetStation.MainStands.Availabilities.ElectricalBikes
		meca := targetStation.MainStands.Availabilities.MechanicalBikes

		hasElec := elec > 0
		hasMeca := meca > 0

		state, exists := m.states[stationID]
		if !exists {
			m.states[stationID] = &StationState{
				HasElec: hasElec,
				HasMeca: hasMeca,
			}
			log.Printf("Initial state for station %d - Elec: %v, Meca: %v", stationID, hasElec, hasMeca)
			continue
		}

		// Check transitions
		if state.HasElec && !hasElec {
			m.notifyAll(chatIDs, fmt.Sprintf("🔴 No more electric Velo'v at station %s (%d).", targetStation.Name, stationID))
		} else if !state.HasElec && hasElec {
			m.notifyAll(chatIDs, fmt.Sprintf("🟢 Electric Velo'v available at station %s (%d)!", targetStation.Name, stationID))
		}

		if state.HasMeca && !hasMeca {
			m.notifyAll(chatIDs, fmt.Sprintf("🔴 No more mechanical Velo'v at station %s (%d).", targetStation.Name, stationID))
		} else if !state.HasMeca && hasMeca {
			m.notifyAll(chatIDs, fmt.Sprintf("🟢 Mechanical Velo'v available at station %s (%d)!", targetStation.Name, stationID))
		}

		// Update state
		state.HasElec = hasElec
		state.HasMeca = hasMeca
	}
}

func (m *Monitor) notifyAll(chatIDs []int64, msg string) {
	log.Println("Broadcasting:", msg)
	for _, chatID := range chatIDs {
		if err := m.notifier.SendTo(chatID, msg); err != nil {
			log.Printf("Failed to send telegram message to %d: %v", chatID, err)
		}
	}
}
