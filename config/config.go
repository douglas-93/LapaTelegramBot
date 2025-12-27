package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Load carrega as variáveis de ambiente do arquivo .env
func Load() {
	// Carrega .env se existir, se não assume que as variáveis já estão no ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente do sistema.")
	}
}

// Get retorna o valor da variável de ambiente ou um valor padrão
func Get(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
