package main

import (
	"LapaTelegramBot/config"
	bot "LapaTelegramBot/telegram"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	// Rodar o bot em uma goroutine para o Windows saber que o serviço "subiu".
	go p.run()
	return nil
}

func (p *program) run() {
	// Para não falhar ao ler env
	// 1. Pega o caminho do executável
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Erro ao obter caminho do exe:", err)
	}

	// 2. Define o diretório de trabalho sendo a pasta onde o exe está
	dir := filepath.Dir(exePath)
	err = os.Chdir(dir)
	if err != nil {
		log.Fatal("Erro ao mudar diretório:", err)
	}

	// 3. Carrega as configurações e inicia o bot no diretório correto
	config.Load()
	bot.StartBot()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "LapaTelegramBot",
		DisplayName: "Lapa Telegram Bot Service",
		Description: "Serviço do Bot do Telegram da Lapa para gerenciamento de serviços e monitoramento",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		err := service.Control(s, os.Args[1])
		if err != nil {
			log.Printf("Erro ao executar comando: %s", err)
		}
		return
	}

	// s.Run() detecta automaticamente se é terminal (Interativo) ou Serviço
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
