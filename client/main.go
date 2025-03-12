package main

import (
	"log"

	"github.com/beliaevke/simpleKeeper/client/app"
)

func main() {

	client := app.NewClient()

	if err := client.Run(); err != nil {
		log.Fatal(err)
	}

}
