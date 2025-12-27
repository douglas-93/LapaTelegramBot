package bot

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleFileUpload(update tgbotapi.Update) {
	var fileID string
	var fileName string

	if update.Message.Document != nil {
		fileID = update.Message.Document.FileID
		fileName = update.Message.Document.FileName
	} else if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
		// Pega a maior versão da foto
		photo := update.Message.Photo[len(update.Message.Photo)-1]
		fileID = photo.FileID
		fileName = fmt.Sprintf("photo_%s.jpg", photo.FileID)
	} else if update.Message.Audio != nil {
		fileID = update.Message.Audio.FileID
		fileName = update.Message.Audio.FileName
	} else if update.Message.Video != nil {
		fileID = update.Message.Video.FileID
		fileName = update.Message.Video.FileName
	} else if update.Message.Voice != nil {
		fileID = update.Message.Voice.FileID
		fileName = fmt.Sprintf("voice_%s.ogg", update.Message.Voice.FileID)
	} else {
		return
	}

	// Feedback para o usuário
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⏳ Recebendo arquivo: %s...", fileName))
	tempMsg, _ := b.API.Send(processingMsg)

	// Cria a pasta se não existir
	uploadDir := "uploaded_files"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, 0755)
		if err != nil {
			b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("❌ Erro ao criar pasta de uploads: %v", err))
			return
		}
	}

	// Obtém o link do arquivo via API do Telegram
	fileURL, err := b.API.GetFileDirectURL(fileID)
	if err != nil {
		b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("❌ Erro ao obter URL do arquivo: %v", err))
		return
	}

	// Faz o download do arquivo
	resp, err := http.Get(fileURL)
	if err != nil {
		b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("❌ Erro ao baixar arquivo: %v", err))
		return
	}
	defer resp.Body.Close()

	// Cria o arquivo local
	destPath := filepath.Join(uploadDir, fileName)
	out, err := os.Create(destPath)
	if err != nil {
		b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("❌ Erro ao salvar arquivo localmente: %v", err))
		return
	}
	defer out.Close()

	// Salva o conteúdo
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("❌ Erro ao gravar arquivo: %v", err))
		return
	}

	b.editMessage(update.Message.Chat.ID, tempMsg.MessageID, fmt.Sprintf("✅ Arquivo *%s* salvo com sucesso em `%s`!", fileName, uploadDir))
}

// helper para editar mensagens
func (b *Bot) editMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	b.API.Send(edit)
}
