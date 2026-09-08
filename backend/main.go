package main

import (
	"log"

	"github.com/AiSiriRak/Artmission/backend/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
