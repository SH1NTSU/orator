package main

import (
	"context"
	"log"

	"github.com/SH1NTSU/orator/application"
)

func main() {
	app := application.New()

	if err := app.Start(context.TODO()); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}
}
