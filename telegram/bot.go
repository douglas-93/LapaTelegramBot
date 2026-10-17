package bot

import (
	"LapaTelegramBot/mailer"
	"LapaTelegramBot/schedule"
	"LapaTelegramBot/zabbix"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API             *tgbotapi.BotAPI
	Zabbix          *zabbix.Client
	Mailer          *mailer.Client
	ScheduleStore   *schedule.Storage
	ScheduleManager *schedule.Manager
	Commands        map[string]func(tgbotapi.Update)
	AllowedChats    map[int64]bool
	Monitors        map[int64]*Monitor
}

func StartBot() {
	token := os.Getenv("TELEGRAM_API_TOKEN")
	chatsIds := os.Getenv("TELEGRAM_ALLOWED_CHAT_ID")

	allowedChatID := strings.Split(chatsIds, ",")
	allowed := loadAllowedChats(allowedChatID)

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot := &Bot{
		API:          api,
		Zabbix:       zabbix.NewClient(),
		Mailer:       mailer.NewClient(),
		AllowedChats: allowed,
		Monitors:     make(map[int64]*Monitor),
	}

	bot.initCommands()
	bot.initSchedule()

	log.Println("Bot iniciado como:", bot.API.Self.UserName)
	bot.Start()
}

func (b *Bot) initSchedule() {
	b.ScheduleStore = schedule.NewStorage()
	b.ScheduleManager = schedule.NewManager()

	b.ScheduleStore.Load()

	// Passa o método ExecuteCommand como callback
	schedule.LoadExistingJobs(b.ScheduleStore, b.ScheduleManager, b.ExecuteCommand)

	b.ScheduleManager.Start()
}

func (b *Bot) initCommands() {
	b.Commands = map[string]func(tgbotapi.Update){
		"status_check":      b.handleStatusCheck,
		"status_monitor":    b.handleStatusMonitor,
		"protheus_status":   b.handleProtheusStatus,
		"listip":            b.handleListIp,
		"ping":              b.handlePing,
		"services":          b.handleRemoteServices,
		"list_services":     b.handleListServices,
		"printers_counter":  b.handlePrinterCounter,
		"schedule_add":      b.handleScheduleAdd,
		"schedule_remove":   b.handleScheduleRemove,
		"schedule_list":     b.handleScheduleList,
		"schedule_help":     b.handleScheduleHelp,
		"restart_win":       b.handleRestartWindowsHost,
		"shutdown_win":      b.handleShutdownWindowsHost,
		"send_mail_counter": b.handleSendMailCounter,
		"resetpwd":          b.handleResetPassword,
	}
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		// Processa callback queries (inline buttons)
		if update.CallbackQuery != nil {
			data := update.CallbackQuery.Data
			// formato esperado: monitor:action:chatID
			parts := strings.Split(data, ":")
			if len(parts) >= 3 && parts[0] == "monitor" {
				action := parts[1]
				chatID := update.CallbackQuery.Message.Chat.ID

				m, ok := b.Monitors[chatID]
				if !ok {
					// responde callback e envia mensagem de erro
					answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "Monitor não encontrado.")
					b.API.Request(answer)
					b.API.Send(tgbotapi.NewMessage(chatID, "Monitor não encontrado para este chat."))
					continue
				}

				switch action {
				case "stop":
					// Remove os botões da última mensagem, se existir
					if m.lastMsgID != 0 {
						edit := tgbotapi.NewEditMessageReplyMarkup(chatID, m.lastMsgID, tgbotapi.InlineKeyboardMarkup{})
						b.API.Send(edit)
						m.lastMsgID = 0
					}
					// sinaliza parada sem fechar o channel diretamente
					select {
					case m.stopCh <- struct{}{}:
					default:
					}
					delete(b.Monitors, chatID)
					// responde callback e envia mensagem ao chat
					answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "Monitor parado.")
					b.API.Request(answer)
					b.API.Send(tgbotapi.NewMessage(chatID, "Monitor parado."))
				case "increase":
					// marca que o monitor está aguardando novo intervalo via mensagem
					m.waitingInterval = true
					answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "Peça enviada.")
					b.API.Request(answer)
					b.API.Send(tgbotapi.NewMessage(chatID, "Envie o novo intervalo em minutos como mensagem nesta conversa."))
				default:
					answer := tgbotapi.NewCallback(update.CallbackQuery.ID, "Ação desconhecida.")
					b.API.Request(answer)
					b.API.Send(tgbotapi.NewMessage(chatID, "Ação desconhecida."))
				}
			}
			continue
		}
		logUpdate(update)
		// Ignora qualquer update sem mensagem
		if update.Message == nil {
			continue
		}

		// Trata comando /start antes de verificar autorização
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			b.handleStart(update)
			continue
		}

		// Verifica se o chat está autorizado
		if !b.AllowedChats[update.Message.Chat.ID] {
			continue
		}

		if !update.Message.IsCommand() { // Processa mensagens não-comando
			// Verifica se é um arquivo (Documento, Foto, Audio, Video, etc)
			if update.Message.Document != nil ||
				update.Message.Photo != nil ||
				update.Message.Audio != nil ||
				update.Message.Video != nil ||
				update.Message.Voice != nil {
				b.handleFileUpload(update)
				continue
			}

			// Se existe um monitor aguardando novo intervalo, trata essa mensagem como novo intervalo
			if m, ok := b.Monitors[update.Message.Chat.ID]; ok && m.waitingInterval {
				text := strings.TrimSpace(update.Message.Text)
				minutes, err := strconv.Atoi(text)
				if err != nil || minutes <= 0 {
					b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Intervalo inválido. Informe um número inteiro de minutos."))
					continue
				}
				// envia novo intervalo para o monitor
				m.waitingInterval = false
				m.updateInterval <- minutes
				b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Intervalo atualizado para %d minutos.", minutes)))
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe um comando.")
			b.API.Send(msg)
			continue
		}

		cmd := update.Message.Command()
		if handler, ok := b.Commands[cmd]; ok {
			handler(update)
		}
	}
}

func (b *Bot) ExecuteCommand(cmd string, chatID int64) {
	// Remove a barra inicial se existir (embora o scheduler geralmente guarde o comando raw)
	cmdClean := strings.TrimPrefix(cmd, "/")
	parts := strings.Split(cmdClean, " ")
	commandName := parts[0]

	// Cria um Update fake para reaproveitar a lógica dos handlers
	fakeUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: "/" + cmdClean, // Recria o texto do comando
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: len(commandName) + 1},
			},
		},
	}

	if handler, ok := b.Commands[commandName]; ok {
		log.Printf("Executando handler via scheduler para comando: %s", commandName)
		handler(fakeUpdate)
	} else {
		log.Printf("Comando não encontrado: %s", commandName)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Comando '%s' não encontrado", commandName))
		b.API.Send(msg)
	}
}

func logUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	name := update.Message.From.FirstName
	if update.Message.From.LastName != "" {
		name += " " + update.Message.From.LastName
	}

	username := update.Message.Chat.UserName
	if username == "" {
		username = "[N/A]"
	}

	comando := update.Message.Text

	log.Printf(
		"\n• Chat ID:%d\n• Nome:%s\n• Username:%s\n• Comando:%s\n\n",
		update.Message.Chat.ID,
		name,
		username,
		comando,
	)
}

func (b *Bot) handleStart(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userName := update.Message.From.FirstName
	if update.Message.From.LastName != "" {
		userName += " " + update.Message.From.LastName
	}

	// Verifica se o usuário está autorizado
	if !b.AllowedChats[chatID] {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"🚫 *Acesso Não Autorizado*\n\n"+
				"Olá, %s!\n\n"+
				"Infelizmente você não tem permissão para usar este bot.\n\n"+
				"Este é um bot privado e apenas usuários autorizados podem utilizá-lo.\n\n"+
				"_Chat ID: %d_",
			userName, chatID,
		))
		msg.ParseMode = "Markdown"
		b.API.Send(msg)
		log.Printf("⚠️  Tentativa de acesso não autorizado - Chat ID: %d, Nome: %s", chatID, userName)
		return
	}

	// Mensagem de boas-vindas para usuários autorizados
	welcomeMsg := fmt.Sprintf(
		"👋 *Bem-vindo, %s!*\n\n"+
			"Sou o *LapaTelegramBot*, seu assistente de gerenciamento e monitoramento.\n\n"+
			"🎯 *Principais Funcionalidades:*\n\n"+
			"🌐 *Monitoramento de Rede*\n"+
			"• `/ping` - Testa conectividade\n"+
			"• `/listip` - Lista hosts do Zabbix\n\n"+
			"📊 *Monitoramento Zabbix*\n"+
			"• `/status_check` - Status dos hosts\n"+
			"• `/printers_counter` - Contadores de impressoras\n"+
			"• `/protheus_status` - Status Protheus/TOTVS\n\n"+
			"⚙️ *Gerenciamento de Serviços*\n"+
			"• `/services` - Gerenciar serviços remotos\n"+
			"• `/list_services` - Listar serviços\n\n"+
			"💻 *Gerenciamento Windows*\n"+
			"• `/restart_win` - Reiniciar host\n"+
			"• `/shutdown_win` - Desligar host\n\n"+
			"🔐 *Active Directory*\n"+
			"• `/resetpwd <usuario>` - Redefinir senha, exigir troca no próximo logon e desbloquear conta\n\n"+
			"📧 *Relatórios*\n"+
			"• `/send_mail_counter` - Enviar contadores por email\n\n"+
			"📁 *Upload de Arquivos*\n"+
			"• Envie qualquer arquivo para salvá-lo no servidor!\n\n"+
			"⏰ *Agendamentos*\n"+
			"• `/schedule_add` - Criar agendamento\n"+
			"• `/schedule_list` - Listar agendamentos\n"+
			"• `/schedule_remove` - Remover agendamento\n"+
			"• `/schedule_help` - Ajuda sobre CRON\n\n"+
			"💡 *Dica:* Todos os comandos fornecem feedback em tempo real!\n\n"+
			"Digite qualquer comando para começar. 🚀",
		userName,
	)

	msg := tgbotapi.NewMessage(chatID, welcomeMsg)
	msg.ParseMode = "Markdown"
	b.API.Send(msg)
}

func loadAllowedChats(parts []string) map[int64]bool {
	allowed := make(map[int64]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Printf("⚠️  Ignorando ID inválido no ALLOWED_CHATS: %s\n", p)
			continue
		}

		allowed[id] = true
	}

	return allowed
}
