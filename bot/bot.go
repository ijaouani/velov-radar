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

		if command == "set" || command == "start" {
			args := strings.TrimSpace(update.Message.CommandArguments())
			if args == "" {
				b.notifier.SendTo(chatID, "❌ Please specify a station, for example: /set 10023")
				continue
			}

			stationID, err := strconv.Atoi(args)
			if err != nil {
				b.notifier.SendTo(chatID, "❌ Invalid station number. Example: /set 10023")
				continue
			}

			b.handleSubscribe(chatID, stationID)
		} else if command == "stop" {
			b.store.Delete(chatID)
			b.notifier.SendTo(chatID, "✅ You have been successfully unsubscribed. You will no longer receive notifications. Use /set <station> to restart.")
			log.Printf("User %d unsubscribed", chatID)
		}
	}
}

func (b *Bot) handleSubscribe(chatID int64, stationID int) {
	targetStation, err := b.monitor.GetStation(stationID)
	if err != nil {
		if err.Error() == "station not found" {
			b.notifier.SendTo(chatID, "❌ Station "+strconv.Itoa(stationID)+" not found. Please check the number.")
		} else {
			// API down fallback
			b.store.Set(chatID, stationID)
			b.notifier.SendTo(chatID, "✅ Great, I am now monitoring station "+strconv.Itoa(stationID)+"! (Current status unavailable)")
			log.Printf("User %d subscribed to station %d without current status", chatID, stationID)
		}
		return
	}

	elec := targetStation.MainStands.Availabilities.ElectricalBikes
	meca := targetStation.MainStands.Availabilities.MechanicalBikes

	b.store.Set(chatID, stationID)
	msg := fmt.Sprintf("✅ Great, I am now monitoring station %s (%d)!\n\nCurrent availability:\n⚡ Electric: %d\n🚲 Mechanical: %d", targetStation.Name, stationID, elec, meca)
	b.notifier.SendTo(chatID, msg)
	log.Printf("User %d subscribed to station %d", chatID, stationID)
}
