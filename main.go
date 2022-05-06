package main

import (
	"github.com/plumber-cd/chcontainer/cmd"
	"github.com/plumber-cd/chcontainer/log"
)

func main() {
	loggerCallback := log.SetupLog()
	defer loggerCallback()

	log.Debug.Print("ChContainer started")

	err := cmd.Execute()
	if err != nil {
		log.Error.Fatal(err)
	}
}
