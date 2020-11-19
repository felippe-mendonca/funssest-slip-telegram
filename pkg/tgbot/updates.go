package tgbot

import (
	"fmt"
	"funssest-slip-telegram/pkg/funssest"

	"github.com/Nhanderu/brdoc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	requestCPFMessage      = "Responda com o número do CPF"
	invalidCPFMessage      = "CPF inválido, tente novamente /boletos."
	failureMessage         = "Ocorreu um problema, tente novamente /boletos."
	noSlipFoundMessage     = "Nenhum boleto em aberto encontrado para o CPF %s."
	commandNotFoundMessage = "Comando não identificado."
	helpCommandMessage     = "Digite /boletos para buscar por CPF."
	failedSendSlipMessage  = "Ocorreu um erro ao enviar um dos boletos, tente novamente /boletos."
)

const (
	helpCommand = "help"
	slipCommand = "boletos"
)

const (
	slipNumberButton = "Número"
	slipURLButton    = "Boleto"
)

func isCPFRequest(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	if msg.ReplyToMessage == nil {
		return false
	}
	return msg.ReplyToMessage.Text == requestCPFMessage && msg.ReplyToMessage.From.IsBot
}

func isCommand(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	return msg.IsCommand()
}

func makeSlipMessageKeyboard(slip funssest.FunssestSlip) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(slipNumberButton, slip.Number),
			tgbotapi.NewInlineKeyboardButtonURL(slipURLButton, slip.URL),
		),
	)
}

func sendMessage(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	_, err := bot.Send(msg)
	if err != nil {
		// log error
	}
	return err
}

func replySlipNumber(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, data))
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, data)
	sendMessage(bot, msg)
}

func processCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	switch update.Message.Command() {
	case helpCommand:
		msg.Text = helpCommandMessage
	case slipCommand:
		msg.Text = requestCPFMessage
		msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true, Selective: false}
	default:
		msg.Text = commandNotFoundMessage
	}

	sendMessage(bot, msg)
}

func processCPFRequest(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	chatID := update.Message.Chat.ID
	cpf := update.Message.Text

	if !brdoc.IsCPF(cpf) {
		sendMessage(bot, tgbotapi.NewMessage(chatID, invalidCPFMessage))
		return
	}

	slips, err := funssest.GetSlips(cpf)
	if err != nil {
		sendMessage(bot, tgbotapi.NewMessage(chatID, failureMessage))
		return
	}

	if len(slips) == 0 {
		sendMessage(bot, tgbotapi.NewMessage(chatID, fmt.Sprintf(noSlipFoundMessage, cpf)))
		return
	}

	for n, slip := range slips {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("<b>#%d</b>", n+1)+slip.Markdown())
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = makeSlipMessageKeyboard(slip)
		if err := sendMessage(bot, msg); err != nil {
			errMsg := tgbotapi.NewMessage(chatID, failedSendSlipMessage)
			defer sendMessage(bot, errMsg)
		}
	}
}

func ProcessUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		replySlipNumber(bot, update)
	} else if isCommand(update.Message) {
		processCommand(bot, update)
	} else if isCPFRequest(update.Message) {
		processCPFRequest(bot, update)
	}
}
