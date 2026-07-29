package main

import (
	"log"

	"fedcm-demo/internal/idp"
)

func main() {
	db, err := idp.OpenDB("data/idp.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	srv := idp.NewServer(db)
	if err := srv.ListenAndServe(":8080"); err != nil {
		log.Fatal(err)
	}
}
