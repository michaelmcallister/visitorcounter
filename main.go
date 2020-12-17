package main

import (
	"flag"
	"log"

	"github.com/michaelmcallister/visitorcounter/datastore"
	"github.com/michaelmcallister/visitorcounter/server"
	"github.com/michaelmcallister/visitorcounter/visitorcounter"
)

var (
	dbLocation = flag.String("db", "./counter.db", "database file location")
	serverAddr = flag.String("http", "127.0.0.1:8080", "HTTP listen address")
)

func main() {
	flag.Parse()

	db, err := datastore.NewBolt(*dbLocation, nil)
	if err != nil {
		log.Fatal("Unable to create database: ", err)
	}
	log.Print("Starting a server on ", *serverAddr)
	r := visitorcounter.NewRender(db)
	s := server.NewServer(r)
	if err := s.ListenAndServe(*serverAddr); err != nil {
		log.Fatal("Unable to start server: ", err)
	}
}
