package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(fmt.Printf("Erro ao carregar o arquivo .env:%v", err))
		return err
	}
	return nil
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}
