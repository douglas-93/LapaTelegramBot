package bot

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const resetADAccountScript = `
$ErrorActionPreference = 'Stop'
try {
    Import-Module ActiveDirectory -ErrorAction Stop
    $user = Get-ADUser -Identity $env:AD_RESET_USERNAME -Properties LockedOut, PasswordNeverExpires -ErrorAction Stop
    $password = ConvertTo-SecureString -String $env:AD_DEFAULT_PASSWORD -AsPlainText -Force
    Set-ADAccountPassword -Identity $user -Reset -NewPassword $password -ErrorAction Stop

    if ($user.PasswordNeverExpires) {
        Set-ADUser -Identity $user -PasswordNeverExpires $false -ErrorAction Stop
    }
    Set-ADUser -Identity $user -ChangePasswordAtLogon $true -ErrorAction Stop

    if ($user.LockedOut) {
        Unlock-ADAccount -Identity $user -ErrorAction Stop
        Write-Output 'UNLOCKED'
    } else {
        Write-Output 'NOT_LOCKED'
    }
} catch {
    [Console]::Error.WriteLine($_.Exception.Message)
    exit 1
}
`

func (b *Bot) handleResetPassword(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	parts := strings.Fields(update.Message.Text)
	if len(parts) != 2 {
		b.API.Send(tgbotapi.NewMessage(chatID, "Uso: /resetpwd <usuario>"))
		return
	}

	defaultPassword := os.Getenv("AD_DEFAULT_PASSWORD")
	if defaultPassword == "" {
		b.API.Send(tgbotapi.NewMessage(chatID, "Erro: AD_DEFAULT_PASSWORD não está configurada."))
		return
	}

	username := parts[1]
	cmd := exec.Command(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		resetADAccountScript,
	)
	cmd.Env = append(os.Environ(),
		"AD_RESET_USERNAME="+username,
		"AD_DEFAULT_PASSWORD="+defaultPassword,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		details := strings.TrimSpace(string(output))
		log.Printf("Falha ao redefinir senha do usuário %q no AD: %v; saída: %s", username, err, details)
		message := fmt.Sprintf("❌ Não foi possível redefinir a senha de %s.", username)
		if details != "" {
			message += "\nErro: " + details
		}
		b.API.Send(tgbotapi.NewMessage(chatID, message))
		return
	}

	switch strings.TrimSpace(string(output)) {
	case "UNLOCKED":
		b.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Senha redefinida para %s, troca obrigatória no próximo logon configurada e conta desbloqueada.", username)))
	case "NOT_LOCKED":
		b.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Senha redefinida para %s e troca obrigatória no próximo logon configurada. A conta não estava bloqueada.", username)))
	default:
		log.Printf("Resposta inesperada ao redefinir senha do usuário %q no AD: %q", username, strings.TrimSpace(string(output)))
		b.API.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Não foi possível confirmar a redefinição de senha para %s.", username)))
	}
}
