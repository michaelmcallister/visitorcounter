package main

import (
	"log"

	"github.com/michaelmcallister/visitorcounter/datastore"
	"github.com/michaelmcallister/visitorcounter/server"
	"github.com/michaelmcallister/visitorcounter/visitorcounter"
)

func main() {
	db, err := datastore.NewBolt("counter.db", nil)
	if err != nil {
		log.Fatalf("Unable to create database: %v\n", err)
	}
	renderer := visitorcounter.NewRender(db)
	s, err := server.NewServer(renderer)
	if err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
	log.Print("Starting a server on :8080")
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Unable to start server: %v\n", err)
	}
}
