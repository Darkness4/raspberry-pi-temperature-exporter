package main

import (
	"log"

	"github.com/Darkness4/raspberry-pi-temperature-exporter/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln("couldn't execute command", err)
	}
}
