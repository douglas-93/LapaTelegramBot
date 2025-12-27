package bot

import (
	"LapaTelegramBot/schedule"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleScheduleAdd(update tgbotapi.Update) {
	parts := strings.Fields(update.Message.Text)

	if len(parts) < 7 {
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
			"Uso: /schedule_add <min> <hora> <dia-mes> <mes> <dia-semana> <comando>"))
		return
	}

	cmd := strings.Replace(parts[6], "/", "", 1)
	if _, exists := b.Commands[cmd]; !exists {
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Verifique se o comando informado é válido para o bot."))
		return
	}

	// cron tem 5 campos
	cronExpr := strings.Join(parts[1:6], " ")
	command := parts[6]
	chatID := update.Message.Chat.ID

	if err := schedule.ValidateCron(cronExpr); err != nil {
		b.API.Send(tgbotapi.NewMessage(chatID, "Erro: "+err.Error()))
		return
	}

	id := time.Now().UnixNano()

	j := schedule.Job{
		ID:      id,
		Cron:    cronExpr,
		Command: command,
		ChatID:  chatID,
		Name:    "Agendamento criado pelo usuário",
	}

	err := b.ScheduleStore.Add(j)
	if err != nil {
		b.API.Send(tgbotapi.NewMessage(chatID, "Erro: "+err.Error()))
		return
	}
	err = b.ScheduleManager.Add(j, func() {
		log.Printf("Executando job agendado: %s (ID: %d)", command, id)
		b.ExecuteCommand(command, chatID)
	})
	if err != nil {
		b.API.Send(tgbotapi.NewMessage(chatID, "Erro: "+err.Error()))
		return
	}

	b.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Agendamento criado! ID: %d", id)))
}

func (b *Bot) handleScheduleRemove(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) != 2 {
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Uso: /schedule_remove <ID>"))
		return
	}

	id, _ := strconv.ParseInt(parts[1], 10, 64)

	b.ScheduleManager.Remove(id)
	b.ScheduleStore.Delete(id)

	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Agendamento removido!"))
}

func (b *Bot) handleScheduleList(update tgbotapi.Update) {
	jobs := b.ScheduleStore.All()

	if len(jobs) == 0 {
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Nenhum agendamento configurado."))
		return
	}

	msg := "📅📅📅 Agendamentos atuais: 📅📅📅\n\n"

	for _, j := range jobs {
		msg += fmt.Sprintf("• *ID:* %d\nCron: `%s`\nCmd: `%s`\n\n",
			j.ID, j.Cron, j.Command)
	}

	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func (b *Bot) handleScheduleHelp(update tgbotapi.Update) {
	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, schedule.CronHelp()))
}
