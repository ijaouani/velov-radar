package bot

import (
	"log"
	"strconv"
	"strings"

	"velov-radar/messages"
	"velov-radar/monitor"
	"velov-radar/notifier"
	"velov-radar/store"
)

type Bot struct {
	notifier *notifier.TelegramNotifier
	store    *store.Store
	monitor  *monitor.Monitor
}

func NewBot(notif *notifier.TelegramNotifier, store *store.Store, mon *monitor.Monitor) *Bot {
	return &Bot{
		notifier: notif,
		store:    store,
		monitor:  mon,
	}
}

func (b *Bot) Listen() {
	updates := b.notifier.GetUpdatesChannel()
	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		chatID := update.Message.Chat.ID
		command := update.Message.Command()
		args := strings.TrimSpace(update.Message.CommandArguments())

		if command == "add" || command == "start" {
			if command == "start" {
				b.store.Clear(chatID)
			}

			if args == "" {
				if command == "start" {
					b.notifier.SendTo(chatID, messages.MsgWelcome)
				} else {
					b.notifier.SendTo(chatID, messages.ErrAddMissingStation)
				}
				continue
			}

			for _, arg := range strings.Fields(args) {
				stationID, err := strconv.Atoi(arg)
				if err != nil {
					b.notifier.SendTo(chatID, messages.ErrAddInvalidStation)
					continue
				}
				b.handleAdd(chatID, stationID)
			}
		} else if command == "remove" {
			if args == "" {
				b.notifier.SendTo(chatID, messages.ErrRemoveMissingStation)
				continue
			}

			for _, arg := range strings.Fields(args) {
				stationID, err := strconv.Atoi(arg)
				if err != nil {
					b.notifier.SendTo(chatID, messages.ErrRemoveInvalidStation)
					continue
				}
				b.store.Remove(chatID, stationID)
				b.notifier.SendTo(chatID, messages.MsgStationRemoved(stationID))
				log.Printf("User %d removed station %d", chatID, stationID)
			}
		} else if command == "list" {
			b.handleList(chatID)
		} else if command == "stop" {
			b.store.Clear(chatID)
			b.notifier.SendTo(chatID, messages.MsgUnsubscribedAll)
			log.Printf("User %d unsubscribed from all stations", chatID)
		}
	}
}

func (b *Bot) handleAdd(chatID int64, stationID int) {
	targetStation, err := b.monitor.GetStation(stationID)
	if err != nil {
		if err.Error() == "station not found" {
			b.notifier.SendTo(chatID, messages.ErrStationNotFound(stationID))
		} else {
			// API down fallback
			b.store.Add(chatID, stationID)
			b.notifier.SendTo(chatID, messages.MsgMonitoringWithoutStatus(stationID))
			log.Printf("User %d subscribed to station %d without current status", chatID, stationID)
		}
		return
	}

	elec := targetStation.MainStands.Availabilities.ElectricalBikes
	meca := targetStation.MainStands.Availabilities.MechanicalBikes

	b.store.Add(chatID, stationID)
	b.notifier.SendTo(chatID, messages.MsgMonitoringWithStatus(targetStation.Name, elec, meca))
	log.Printf("User %d subscribed to station %d", chatID, stationID)
}

func (b *Bot) handleList(chatID int64) {
	subs := b.store.GetAll()
	stations, ok := subs[chatID]
	if !ok || len(stations) == 0 {
		b.notifier.SendTo(chatID, messages.MsgNoMonitoredStations)
		return
	}

	var lines []string
	lines = append(lines, messages.MsgListHeader)
	for _, stationID := range stations {
		st, err := b.monitor.GetStation(stationID)
		if err != nil {
			lines = append(lines, messages.MsgListStationUnavailable(stationID))
		} else {
			elec := st.MainStands.Availabilities.ElectricalBikes
			meca := st.MainStands.Availabilities.MechanicalBikes
			lines = append(lines, messages.MsgListStation(st.Name, elec, meca))
		}
	}

	b.notifier.SendTo(chatID, strings.Join(lines, "\n"))
}
