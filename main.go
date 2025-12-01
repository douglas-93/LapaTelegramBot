package main

import (
	"TelegramNotify/config"
	bot "TelegramNotify/telegram"
)

func main() {
	env := config.LoadConfig()
	bot.StartBot(env)
}
