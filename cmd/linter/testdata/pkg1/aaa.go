package pkg1

import (
	"log"
	"os"
)

func hello() {
	panic("boom") // want "panic function call!"
}

func bye() {
	os.Exit(1) // want "log.Fatal or os.Exit calls outside of main function are prohibited!"
}

func fatal() {
	log.Fatal("stop") // want "log.Fatal or os.Exit calls outside of main function are prohibited!"
}
