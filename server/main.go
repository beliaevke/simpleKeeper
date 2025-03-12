package main

import (
	"log"

	"github.com/beliaevke/simpleKeeper/server/app"

	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {

	srv := app.NewServer()

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}

}
