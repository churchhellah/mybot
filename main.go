package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NetAPI("8219522776:AAG2wlgT8JPHTwBUD9wYvCoBnT2AqOVvomk")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			var msg tgbotapi.MessageConfig

			switch update.Message.Command() {
			case "start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome! I am your bot.")
			case "help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "I can help you with the following commands:\n/start - Start the bot\n/help - Display this help message")
			default:
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command")
			}

			bot.Send(msg)
		}
	}
}
