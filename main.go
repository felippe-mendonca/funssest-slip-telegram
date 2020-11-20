package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"funssest-slip-telegram/pkg/tgbot"
)

func main() {
	telegramAccessToken := os.Getenv("TELEGRAM_ACCESS_TOKEN")
	telegramWebhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	telegramDebug, _ := strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))
	httpListenServer := os.Getenv("HTTP_LISTEN_SERVER")

	bot, err := tgbotapi.NewBotAPI(telegramAccessToken)
	if err != nil {
		panic(err)
	}
	bot.Debug = telegramDebug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(telegramWebhookURL + "/updates"))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	updates := bot.ListenForWebhook("/updates")
	go http.ListenAndServe(httpListenServer, nil)

	for update := range updates {
		tgbot.ProcessUpdate(bot, update)
	}
}
