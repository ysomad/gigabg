package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/server"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	cardStore, err := cards.New()
	if err != nil {
		log.Fatal(err)
	}

	memstore := server.NewMemoryStore(cardStore)

	// Pre-create lobby "1" for testing
	if err := memstore.Create("1", nil); err != nil {
		log.Fatal(err)
	}
	log.Println("created lobby '1'")

	srv := &http.Server{
		Addr:         *addr,
		Handler:      server.New(memstore),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("starting server on %s", *addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
