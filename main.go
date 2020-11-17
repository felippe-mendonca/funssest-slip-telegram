package main

import (
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"funssest-slip-telegram/pkg/tgbot"
)

func main() {
	accessToken := os.Getenv("TELEGRAM_ACCESS_TOKEN")
	telegramDebug, _ := strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))

	bot, err := tgbotapi.NewBotAPI(accessToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = telegramDebug
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		tgbot.ProcessUpdate(bot, update)
	}
}
