package main

import (
	"log"

	"fedcm-101/internal/sp"
)

func main() {
	srv := sp.NewServer()
	if err := srv.ListenAndServe(":8081"); err != nil {
		log.Fatal(err)
	}
}
