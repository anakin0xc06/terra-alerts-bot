package main

import (
	"encoding/json"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/anakin0xc06/terra-alerts-bot/config"
	"github.com/anakin0xc06/terra-alerts-bot/helpers"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)
var subscribers = make(map[string][]string)
var validatorsMissedVotes = make(map[string]int64)

func initBot() {
	jsonFile, err := os.Open(config.SubscribersFile)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &subscribers)
	fmt.Println(subscribers)
	validatorsfile, err := os.Open(config.ValidatorsFile)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer validatorsfile.Close()
	byteValue2, _ := ioutil.ReadAll(validatorsfile)
	json.Unmarshal(byteValue2, &validatorsMissedVotes)
	UpdateValidatorMissedVotes()
}


func main() {
	bot, err := tgbotapi.NewBotAPI(config.BOT_API_KEY)
	if err != nil {
		log.Fatalf("Error in instantiating the bot: %v", err)
	}
	initBot()
	go SubscribersHandleScheduler(bot)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		color.Red("Error while receiving messages: %s", err)
		return
	}
	color.Green("Started %s successfully", bot.Self.UserName)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand()  {
			MainHandler(bot, update)
		}
	}
}

// MainHandler ...
func MainHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {

	if update.Message != nil && update.Message.IsCommand() && update.Message.Chat.IsPrivate(){
		command := update.Message.Command()

		switch command {
		case "start":
			text := "Welcome to terra-alerts bot\n"
			helpers.SendMessage(bot, update, text, "html")
		case "subscribe":
			HandleSubscribe(bot, update)
		default:
			text := "Command not available"
			fmt.Println(command, text)
			// helpers.SendMessage(bot, update, text, "html")
		}
	}
}

func HandleSubscribe(bot *tgbotapi.BotAPI, update tgbotapi.Update)  {
	args := update.Message.CommandArguments()
	var validatorAddresses  []string

	if len(args) > 0 {
		arguments := strings.Split(args, " ")
		for _, arg := range arguments {
			if isCorrectValAddress(arg) && !contains(validatorAddresses, arg){
				validatorAddresses = append(validatorAddresses, arg)
			}
		}
		if len(validatorAddresses) > 0 {
			userId := helpers.GetUserID(update)
			validators, ok := subscribers[fmt.Sprint(userId)]
			if !ok {
				subscribers[fmt.Sprint(userId)] = validatorAddresses
			} else {
				for _, val := range validatorAddresses {
					if !contains(validators, val) {
						validators = append(validators, val)
					}
				}
				subscribers[fmt.Sprint(userId)] = validators
			}
			jsonString, _ := json.MarshalIndent(subscribers, "", " ")
			_ = ioutil.WriteFile(config.SubscribersFile, jsonString, 0644)
			helpers.SendMessage(bot, update, "subscribed to alerts.", tgbotapi.ModeHTML)
			return
		} else {
			helpers.SendMessage(bot, update, "Invalid args", tgbotapi.ModeHTML)
			return
		}

	} else {
		helpers.SendMessage(bot, update, "Invalid format, Please use /subscribe [validator addresses ..]", tgbotapi.ModeHTML)
		return
	}
}
func UpdateValidatorMissedVotes()  {
	var addresses []string
	for _, validators := range subscribers {
		for _, validator := range validators {
			if !contains(addresses, validator) {
				addresses = append(addresses, validator)
			}
		}
	}
	for _, address := range addresses {
		votesCount, err := helpers.CheckOracleMissedVotes(address)
		if err != nil {
			continue
		}
		validatorsMissedVotes[address] = votesCount
	}
	validatorsData, _ := json.MarshalIndent(validatorsMissedVotes, "", " ")
	_ = ioutil.WriteFile(config.ValidatorsFile, validatorsData, 0644)
	fmt.Println("Updated validators missed votes data")
}

func HandleSubscribers(bot *tgbotapi.BotAPI)  {
	for user, validators := range subscribers {
		for _, validator := range validators {
			currentMissedVotes, err := helpers.CheckOracleMissedVotes(validator)
			if err != nil {
				continue
			}
			previousMissedVotes, ok := validatorsMissedVotes[validator]
			if !ok {
				continue
			}
			if currentMissedVotes - previousMissedVotes > 0 {
				fmt.Println(validator, "missing oracle votes")
				text := fmt.Sprintf("**Alert**:\n\n %s is missing oracle votes MissedVotesCount **%d -> %d**",
					validator, previousMissedVotes, currentMissedVotes)
				userId, _ := strconv.ParseInt(user, 10, 64)
				msg := tgbotapi.NewMessage(userId, text)
				msg.ParseMode = tgbotapi.ModeMarkdown
				bot.Send(msg)
			}
		}
	}
	UpdateValidatorMissedVotes()
}

func isCorrectValAddress(address string) bool {
	if strings.HasPrefix(address, "terravaloper1") && len(address) == 51 {
		return true
	}
	return false
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func SubscribersHandleScheduler(bot *tgbotapi.BotAPI) {
	go HandleSubscribers(bot)
	s := gocron.NewScheduler()
	s.Every(60).Seconds().Do(HandleSubscribers, bot)
	<-s.Start()
}
