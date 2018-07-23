package main

import (
	"net/http"
	"log"
	"os"
	"./core"
)

func main() {
	var registry = &core.Registry{}
	registry.InitRegistry()

	http.HandleFunc("/register", registry.RegisterHandler)
	http.HandleFunc("/services", registry.ServicesHandler)
	registry.ManageHealthcheckers()
	log.Fatal(http.ListenAndServe(":" + os.Getenv("SR_PORT"), nil))
}