package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"log"
)

func main() {
	handler := handlers.NewHandler()
	myServer := new(server.Server)
	err := myServer.ServerRun(handler.Init(), "8080")
	log.Fatal(err)
}
