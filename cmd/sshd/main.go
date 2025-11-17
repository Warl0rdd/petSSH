package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"petssh/internal/server"
	"syscall"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <listen-address>\n", os.Args[0])
	}

	addr := os.Args[1]

	s, err := server.New(addr)

	if err != nil {
		log.Fatal(err)
	}

	if err := s.ListenAndServe(nil); err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigCh
		fmt.Println("Shutting down...")
		close(done)
	}()

	if err := s.ListenAndServe(done); err != nil {
		log.Fatal(err)
	}
}
