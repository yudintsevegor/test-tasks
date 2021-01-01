package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"not-for-work/aviasales_test/cache"
	"not-for-work/aviasales_test/config"
	"not-for-work/aviasales_test/server"
	"not-for-work/aviasales_test/store"
)

func main() {
	cfg, err := config.Load("./config/local.env")
	if err != nil {
		log.Panic(err)
	}
	log.Printf("CONFIG: %+v", cfg)

	c, err := cache.New(cfg.Redis, cfg.DropData)
	if err != nil {
		log.Panic(err)
	}
	defer c.Close()

	s, err := store.New(cfg.Mongo, cfg.DropData)
	if err != nil {
		log.Panic(err)
	}
	defer s.Close(context.Background())

	go func() {
		log.Printf("start server on port %q", cfg.Port)
		err = http.ListenAndServe(":"+cfg.Port, server.New(c, s))
		if err != nil {
			log.Panic(err)
		}
	}()

	osSigChan := make(chan os.Signal, 1)
	signal.Notify(osSigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-osSigChan
	log.Println("OS interrupting signal has received")
}
