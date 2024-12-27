package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/duziem/ecommerce_proj/cmd/api"
	"github.com/duziem/ecommerce_proj/configs"
	"github.com/duziem/ecommerce_proj/db"
	_ "github.com/lib/pq"
)

func main() {
	cfg := db.PostgresConfig{
		Host:     configs.Envs.DBHost,
		Port:     configs.Envs.DBPort,
		User:     configs.Envs.DBUser,
		Password: configs.Envs.DBPassword,
		DbName:   configs.Envs.DbName,
		SSLMode:  "disable",
	}

	db, err := db.NewPostgresStorage(cfg)
	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)

	defer db.Close()

	server := api.NewAPIServer(fmt.Sprintf(":%s", configs.Envs.Port), db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}
