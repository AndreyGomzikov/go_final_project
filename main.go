package main

import (
	"log"
	"os"
	"strconv"

	"go1f/pkg/db"
	"go1f/pkg/server"
)

const (
	defaultDB   = "scheduler.db"
	defaultPort = 7540
	envDB       = "TODO_DBFILE"
	envPort     = "TODO_PORT"
)

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envPortOr(def int) int {
	if s, ok := os.LookupEnv(envPort); ok && s != "" {
		if p, err := strconv.Atoi(s); err == nil && p > 0 && p <= 65535 {
			return p
		}
	}
	return def
}

func main() {
	dbPath := envOr(envDB, defaultDB)
	if err := db.Open(dbPath); err != nil {
		log.Fatalf("open DB %q: %v", dbPath, err)
	}
	defer db.DB.Close()

	port := envPortOr(defaultPort)
	if err := server.Run(port); err != nil {
		log.Fatalf("start http server on port %d: %v", port, err)
	}
}
