package main

import (
	"Go-api/pkg"
	"log"
)

func main() {
	app := pkg.NewApp()
	if err := app.Run(":8080"); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}