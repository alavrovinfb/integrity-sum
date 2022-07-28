package configs

import (
	"log"
	"os"
)

func ValidateDBConnectionValues() {
	DbDriver, ok := os.LookupEnv("DB_DRIVER")
	if !ok {
		DbDriver = "postgres"
		log.Printf("DB_DRIVER was not set. setting by default: %s", DbDriver)
	}

	DbHost, ok := os.LookupEnv("DB_HOST")
	if !ok {
		DbHost = "localhost"
		log.Printf("DB_HOST was not set. setting by default: %s", DbHost)
	}

	DbPassword, ok := os.LookupEnv("DB_PASSWORD")
	if !ok {
		log.Fatalf("DB_PASSWORD was not set. setting by default: %s", DbPassword)
	}

	DbUser, ok := os.LookupEnv("DB_USER")
	if !ok {
		log.Fatalf("DB_USER was not set. setting by default: %s", DbUser)
	}

	DbPort, ok := os.LookupEnv("DB_PORT")
	if !ok {
		DbPort = "5432"
		log.Printf("DB_PORT was not set. setting by default: %s", DbPort)
	}

	DbName, ok := os.LookupEnv("DB_NAME")
	if !ok {
		log.Fatalf("DB_NAME was not set. setting by default: %s", DbName)
	}

}
