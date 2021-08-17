package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"gopkg.in/telegram-bot-api.v4"
	"strconv"

)
const (
	LCD_URL = "https://lcd.terra.dev"
)

type MissedVotesResponse struct {
	Height string `json:"height"`
	Result string `json:"result"`
}

// GetUserName ...
func GetUserName(u tgbotapi.Update) string {
	var username string
	if u.CallbackQuery != nil {
		username = u.CallbackQuery.From.UserName
	}
	if u.Message != nil {
		username = u.Message.From.UserName
	}
	return username
}

// GetChatID ...
func GetChatID(u tgbotapi.Update) int64 {
	var chatID int64
	if u.CallbackQuery != nil {
		chatID = u.CallbackQuery.Message.Chat.ID
	}
	if u.Message != nil {
		chatID = u.Message.Chat.ID
	}
	return chatID
}

// GetUserID ...
func GetUserID(u tgbotapi.Update) int {
	var userID int
	if u.CallbackQuery != nil {
		userID = u.CallbackQuery.From.ID
	}
	if u.Message != nil {
		userID = u.Message.From.ID
	}
	return userID
}

//GetMsgID ...
func GetMsgID(u tgbotapi.Update) int {
	var MsgID int
	if u.CallbackQuery != nil {
		MsgID = u.CallbackQuery.Message.MessageID
	}
	if u.Message != nil {
		MsgID = u.Message.MessageID
	}
	return MsgID
}

// SendMessage ...
func SendMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string,
	mode string, btns ...tgbotapi.InlineKeyboardMarkup) {

	if update.Message != nil {
		msg := tgbotapi.NewMessage(GetChatID(update), text)
		if len(btns) > 0 {
			msg.ReplyMarkup = btns[0]
		}
		msg.ParseMode = tgbotapi.ModeMarkdown
		if mode != "" {
			msg.ParseMode = mode
		}
		bot.Send(msg)
		return
	}
	if len(btns) > 0 {
		msg := tgbotapi.NewEditMessageText(GetChatID(update), GetMsgID(update), text)
		msg.ReplyMarkup = &btns[0]
		msg.ParseMode = tgbotapi.ModeMarkdown
		if mode != "" {
			msg.ParseMode = mode
		}
		bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(GetChatID(update), text)
	msg.ParseMode = mode
	bot.Send(msg)
	return
}

// SendMessage ...
func SendReplyMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, text string,
	mode string, btns ...tgbotapi.InlineKeyboardMarkup) {

	if update.Message != nil {
		msg := tgbotapi.NewMessage(GetChatID(update), text)
		if len(btns) > 0 {
			msg.ReplyMarkup = btns[0]
		}
		msg.ParseMode = tgbotapi.ModeMarkdown
		if mode != "" {
			msg.ParseMode = mode
		}
		msg.ReplyToMessageID = GetMsgID(update)
		bot.Send(msg)
		return
	}
	if len(btns) > 0 {
		msg := tgbotapi.NewEditMessageText(GetChatID(update), GetMsgID(update), text)
		msg.ReplyMarkup = &btns[0]
		msg.ParseMode = tgbotapi.ModeMarkdown
		if mode != "" {
			msg.ParseMode = mode
		}
		bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(GetChatID(update), text)
	msg.ParseMode = mode
	bot.Send(msg)
	return
}

func CheckOracleMissedVotes(validatorAddress string) (int64, error) {
	url := LCD_URL + fmt.Sprintf("/oracle/voters/%s/miss", validatorAddress)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	defer resp.Body.Close()
    var body MissedVotesResponse
	fmt.Println("response Status:", resp.Status)
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, err
	}
	if resp.Status == "200 OK" {
		missedvotes, err := strconv.ParseInt(body.Result, 10, 64)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		return missedvotes, nil

	}
	return 0, fmt.Errorf("unable to get missed votes count")
}

