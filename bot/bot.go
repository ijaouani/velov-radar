package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
			if args == "" {
				b.notifier.SendTo(chatID, "❌ Please specify a station, for example: /add 10023")
				continue
			}

			stationID, err := strconv.Atoi(args)
			if err != nil {
				b.notifier.SendTo(chatID, "❌ Invalid station number. Example: /add 10023")
				continue
			}

			b.handleAdd(chatID, stationID)
		} else if command == "remove" {
			if args == "" {
				b.notifier.SendTo(chatID, "❌ Please specify a station, for example: /remove 10023")
				continue
			}

			stationID, err := strconv.Atoi(args)
			if err != nil {
				b.notifier.SendTo(chatID, "❌ Invalid station number. Example: /remove 10023")
				continue
			}

			b.store.Remove(chatID, stationID)
			b.notifier.SendTo(chatID, fmt.Sprintf("✅ Station %d removed from your watch list.", stationID))
			log.Printf("User %d removed station %d", chatID, stationID)
		} else if command == "list" {
			b.handleList(chatID)
		} else if command == "stop" {
			b.store.Clear(chatID)
			b.notifier.SendTo(chatID, "✅ You have been successfully unsubscribed from all stations. Use /add <station> to restart.")
			log.Printf("User %d unsubscribed from all stations", chatID)
		}
	}
}

func (b *Bot) handleAdd(chatID int64, stationID int) {
	targetStation, err := b.monitor.GetStation(stationID)
	if err != nil {
		if err.Error() == "station not found" {
			b.notifier.SendTo(chatID, "❌ Station "+strconv.Itoa(stationID)+" not found. Please check the number.")
		} else {
			// API down fallback
			b.store.Add(chatID, stationID)
			b.notifier.SendTo(chatID, "✅ Great, I am now monitoring station "+strconv.Itoa(stationID)+"! (Current status unavailable)")
			log.Printf("User %d subscribed to station %d without current status", chatID, stationID)
		}
		return
	}

	elec := targetStation.MainStands.Availabilities.ElectricalBikes
	meca := targetStation.MainStands.Availabilities.MechanicalBikes

	b.store.Add(chatID, stationID)
	msg := fmt.Sprintf("✅ Great, I am now monitoring station %s (%d)!\n\nCurrent availability:\n⚡ Electric: %d\n🚲 Mechanical: %d", targetStation.Name, stationID, elec, meca)
	b.notifier.SendTo(chatID, msg)
	log.Printf("User %d subscribed to station %d", chatID, stationID)
}

func (b *Bot) handleList(chatID int64) {
	subs := b.store.GetAll()
	stations, ok := subs[chatID]
	if !ok || len(stations) == 0 {
		b.notifier.SendTo(chatID, "You are not monitoring any stations right now. Use /add <station> to add one.")
		return
	}

	var lines []string
	lines = append(lines, "📋 **Your Monitored Stations:**\n")
	for _, stationID := range stations {
		st, err := b.monitor.GetStation(stationID)
		if err != nil {
			lines = append(lines, fmt.Sprintf("• Station %d (status unavailable)", stationID))
		} else {
			elec := st.MainStands.Availabilities.ElectricalBikes
			meca := st.MainStands.Availabilities.MechanicalBikes
			lines = append(lines, fmt.Sprintf("• %s (%d)\n  ⚡ %d | 🚲 %d", st.Name, stationID, elec, meca))
		}
	}

	b.notifier.SendTo(chatID, strings.Join(lines, "\n"))
}
