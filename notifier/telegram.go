package notifier

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramNotifier(token string) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("initializing telegram bot: %w", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot: bot,
	}, nil
}

func (n *TelegramNotifier) SendTo(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := n.bot.Send(msg)
	return err
}

func (n *TelegramNotifier) GetUpdatesChannel() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return n.bot.GetUpdatesChan(u)
}
