package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("stop") // ok
	os.Exit(0)        // ok
}

func helper() {
	os.Exit(0) // want "log.Fatal or os.Exit calls outside of main function are prohibited!"
}
