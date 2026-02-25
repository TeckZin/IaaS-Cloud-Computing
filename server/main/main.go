package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("starting")

	host := os.Getenv("POSTGRES_HOST")

	if host == "" {
		host = "localhost"
	}

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")

	dbName := os.Getenv("POSTGRES_DB")

	fmt.Println(user, pass)

	pgPort := os.Getenv("POSTGRES_PORT")
	if pgPort == "" {
		pgPort = "5432"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, pgPort, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("db ping failed: ", err)
	}

	server := newServer(db)
	handler := server.initRoutes()

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	addr := ":" + httpPort
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
