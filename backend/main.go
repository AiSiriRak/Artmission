package main

import (
	"log"

	"github.com/DeepAung/artmission/backend/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
