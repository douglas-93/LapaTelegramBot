package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadConfig() map[string]string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var env map[string]string
	env, e := godotenv.Read()

	if e != nil {
		log.Fatal(e)
	}

	return env
}
