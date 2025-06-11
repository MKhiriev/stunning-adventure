package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/misc"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"log"
)

func main() {
	misc.ParseServerFlags()
	handler := handlers.NewHandler()
	myServer := new(server.Server)
	err := myServer.ServerRun(handler.Init(), misc.ServerAddress.String())
	log.Fatal(err)
}
