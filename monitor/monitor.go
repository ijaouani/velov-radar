package monitor

import (
	"log"
	"time"

	"velov-radar/messages"
	"velov-radar/notifier"
	"velov-radar/stationcache"
	"velov-radar/store"
)

type StationState struct {
	HasElec bool
	HasMeca bool
}

type Monitor struct {
	cache        *stationcache.Cache
	notifier     *notifier.TelegramNotifier
	store        *store.Store
	pollInterval time.Duration
	states       map[int]*StationState
}

func NewMonitor(cache *stationcache.Cache, notif *notifier.TelegramNotifier, store *store.Store, interval time.Duration) *Monitor {
	return &Monitor{
		cache:        cache,
		notifier:     notif,
		store:        store,
		pollInterval: interval,
		states:       make(map[int]*StationState),
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

func (m *Monitor) check() {
	if err := m.cache.Refresh(); err != nil {
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
		targetStation, err := m.cache.GetStation(stationID)
		if err != nil {
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
			m.notifyAll(chatIDs, messages.NotifyElecEmpty(targetStation.Name))
		} else if !state.HasElec && hasElec {
			m.notifyAll(chatIDs, messages.NotifyElecAvailable(targetStation.Name))
		}

		if state.HasMeca && !hasMeca {
			m.notifyAll(chatIDs, messages.NotifyMecaEmpty(targetStation.Name))
		} else if !state.HasMeca && hasMeca {
			m.notifyAll(chatIDs, messages.NotifyMecaAvailable(targetStation.Name))
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
