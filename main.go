package main

import (
	"log"
	"time"

	"velov-radar/bot"
	"velov-radar/config"
	"velov-radar/monitor"
	"velov-radar/notifier"
	"velov-radar/stationcache"
	"velov-radar/store"
	"velov-radar/velov"
)

func main() {
	log.Println("Starting Velo'v Monitor...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	notif, err := notifier.NewTelegramNotifier(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to initialize notifier: %v", err)
	}

	userStore, err := store.NewStore("users.json")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	client := velov.NewClient()
	cache := stationcache.NewCache(client)

	// Poll every 1 minute
	mon := monitor.NewMonitor(cache, notif, userStore, 60*time.Second)

	log.Printf("Monitoring started...")
	go mon.Start()

	// Initialize the Telegram Bot Handler
	telegramBot := bot.NewBot(notif, userStore, cache)

	// Listen for incoming telegram messages (blocking call)
	telegramBot.Listen()
}
