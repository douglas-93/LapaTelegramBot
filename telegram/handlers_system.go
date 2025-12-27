package bot

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handlePing(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o IP. Ex: /ping 192.168.0.1")
		b.API.Send(msg)
		return
	}

	var wg sync.WaitGroup
	result := make(chan string)

	chatID := update.Message.Chat.ID
	for i := 1; i < len(parts); i++ {
		ip := parts[i]

		wg.Add(1)
		go pingFunc(ip, &wg, result)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	for resultText := range result {
		msg := tgbotapi.NewMessage(chatID, resultText)
		b.API.Send(msg)
	}
}

func pingFunc(ip string, wg *sync.WaitGroup, channel chan string) {
	defer wg.Done()
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		channel <- err.Error()
		return
	}

	pinger.Count = 3
	pinger.Interval = 300 * time.Millisecond
	pinger.Timeout = 3 * time.Second

	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true) /* Falha no Windows caso o programa não seja executado como administrador */
	}

	err = pinger.Run()
	if err != nil {
		// Erro típico de host offline no Windows
		if strings.Contains(strings.ToLower(err.Error()), "wsarecvfrom") {
			channel <- fmt.Sprintf("❌ %s\nStatus: OFFLINE (nenhuma resposta)", ip)
		} else {
			channel <- fmt.Sprintf("❌ %s\nErro no ping: %v", ip, err)
		}
		return
	}

	stats := pinger.Statistics()

	response := fmt.Sprintf(
		"✅ %s\nEnviados: %d | Recebidos: %d | Perda: %.0f%%\nLatência média: %v",
		ip,
		stats.PacketsSent,
		stats.PacketsRecv,
		stats.PacketLoss,
		stats.AvgRtt,
	)
	channel <- response
}

func (b *Bot) handleRestartWindowsHost(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /restart_win \\\\LVMAQUINA")
		b.API.Send(msg)
		return
	}

	host := parts[1]
	log.Printf("Handler restart_win acionado, destino: %s", host)

	cmd := exec.Command(
		"shutdown",
		"/r",
		"/t", "0",
		"/m", fmt.Sprintf("\\\\%s", host),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		e := fmt.Sprintf("Erro ao tentar reiniciar %s: %v\nSaída: %s", host, err, string(output))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, e)
		b.API.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	b.API.Send(msg)
}

func (b *Bot) handleShutdownWindowsHost(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /shutdown_win \\\\LVMAQUINA")
		b.API.Send(msg)
		return
	}
	host := parts[1]
	log.Printf("Handler shutdown_win acionado, destino: %s", host)

	cmd := exec.Command(
		"shutdown",
		"/s",
		"/t", "0",
		"/m", fmt.Sprintf("\\\\%s", host),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		e := fmt.Sprintf("Erro ao tentar desligar %s: %v\nSaída: %s", host, err, string(output))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, e)
		b.API.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	b.API.Send(msg)
}
